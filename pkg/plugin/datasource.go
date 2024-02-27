package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/httpclient"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/wildmountainfarms/wild-graphql-datasource/pkg/plugin/parsing"
	"github.com/wildmountainfarms/wild-graphql-datasource/pkg/plugin/querymodel"
	"github.com/wildmountainfarms/wild-graphql-datasource/pkg/plugin/queryvariables"
	"github.com/wildmountainfarms/wild-graphql-datasource/pkg/util/graphql"
	"net/http"
)

// Make sure Datasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler interfaces. Plugin should not implement all these
// interfaces - only those which are required for a particular task.
var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

// NewDatasource creates a new datasource instance.
func NewDatasource(ctx context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	// https://community.grafana.com/t/how-to-make-user-configurable-http-requests-from-your-data-source-plugin/59724
	//   Since we use the DataSourceHttpSettings, we can easily parse the settings
	httpOptions, err := settings.HTTPClientOptions(ctx)
	if err != nil {
		return nil, err
	}
	client, err := httpclient.New(httpOptions)
	if err != nil {
		return nil, err
	}

	return &Datasource{
		settings:   settings,
		httpClient: client,
	}, nil
}

// Datasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type Datasource struct {
	settings   backend.DataSourceInstanceSettings
	httpClient *http.Client
}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (d *Datasource) Dispose() {
	// Clean up datasource instance resources.
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	response := backend.NewQueryDataResponse()

	// We are currently not implementing any sort of batching strategy.
	//   First off, not every GraphQL server supports batching
	//   More info here: https://github.com/graphql/graphql-spec/issues/375 and also here: https://github.com/graphql/graphql-spec/issues/583#issuecomment-491807207
	// It's absolutely possible for us to try to combine multiple queries into a single one,
	//   but attempting to do that is out of scope for us right now, especially with how complicated a GraphQL query can be.

	for _, q := range req.Queries {
		res, err := d.query(ctx, req.PluginContext, q)
		if err != nil {
			// If an error is returned from the query, we assume that it is not a recoverable error.
			//   We can consider changing this in the future
			return nil, err
		}

		// save the response in a hashmap based on with RefID as identifier
		response.Responses[q.RefID] = *res
	}

	return response, nil
}

func statusFromResponse(response http.Response) backend.Status {
	for _, status := range []backend.Status{} {
		if response.StatusCode == int(status) {
			return status
		}
	}
	return backend.StatusUnknown
}

// Executes a single GraphQL query.
// In most error scenarios, the error should be nested within the DataResponse.
// In some cases that are never expected to happen, error is returned and the DataResponse is nil.
// In these cases, you can assume that something is seriously wrong, as we didn't intend to recover from that specific situation.
func (d *Datasource) query(ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery) (*backend.DataResponse, error) {

	log.DefaultLogger.Info(fmt.Sprintf("JSON is: %s", query.JSON))

	// Unmarshal the JSON into our QueryModel.
	var qm querymodel.QueryModel

	err := json.Unmarshal(query.JSON, &qm)
	if err != nil {
		// A JSON parsing error *could* occur if someone screws up the JSON of a particular query manually.
		//   When that happens, we want to actually prepare for it, even though it's extremely unlikely.
		//   By not returning an error and instead nested it in the DataResponse,
		//   we tell Grafana that the error is within a specific query.
		return &backend.DataResponse{
			Error:  err,
			Status: backend.StatusBadRequest,
		}, nil
	}
	log.DefaultLogger.Info("Query text is: " + qm.QueryText)

	// use later: pCtx.AppInstanceSettings.DecryptedSecureJSONData
	variables, _ := queryvariables.ParseVariables(query, qm.Variables)

	graphQLRequest := graphql.GraphQLRequest{
		Query:         qm.QueryText,
		OperationName: qm.OperationName,
		Variables:     variables,
	}
	request, err := graphQLRequest.ToRequest(ctx, d.settings.URL)
	if err != nil {
		// We don't expect the conversion of the GraphQLRequest into a http.Request to fail
		return nil, err
	}

	resp, err := d.httpClient.Do(request)
	if err != nil {
		// http.Client.Do returns an error when there's a network connectivity problem or something weird going on,
		//   so we expect this to happen every once in a while
		return &backend.DataResponse{
			Error:  err,
			Status: backend.StatusBadRequest,
		}, nil
	}
	status := statusFromResponse(*resp)

	graphQLResponse, responseParseError := graphql.ParseGraphQLResponse(resp.Body)
	if responseParseError != nil {
		return &backend.DataResponse{
			Error:  err,
			Status: status,
		}, nil
	}
	if len(graphQLResponse.Errors) > 0 {
		var errorsString = ""
		for i, graphQLError := range graphQLResponse.Errors {
			if i != 0 {
				errorsString += ", "
			}
			errorsString += graphQLError.Message
		}
		return &backend.DataResponse{
			Error:  errors.New(fmt.Sprintf("GraphQL response had %d error(s): %s", len(graphQLResponse.Errors), errorsString)),
			Status: status,
		}, nil
	}
	if resp.StatusCode != 200 {
		return &backend.DataResponse{
			Error:  errors.New("got non-200 status: " + resp.Status),
			Status: status,
		}, nil
	}

	var response backend.DataResponse

	// add the frames to the response.
	for _, parsingOption := range qm.ParsingOptions {
		frames, err, errorType := parsing.ParseData(
			graphQLResponse.Data,
			parsingOption,
		)
		if err != nil {
			if errorType == parsing.FRIENDLY_ERROR {
				return &backend.DataResponse{
					Error:  err,
					Status: status,
				}, nil
			}
			return nil, err
		}
		response.Frames = append(response.Frames, frames...)
	}

	return &response, nil
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	// test command to do the same thing:
	//   curl -X POST -H "Content-Type: application/json" -d '{"query":"{\n\t\t  __schema{\n\t\t\tqueryType{name}\n\t\t  }\n\t\t}"}' https://swapi-graphql.netlify.app/.netlify/functions/index
	graphQLRequest := graphql.GraphQLRequest{
		Query: `{
		  __schema{
		    queryType{name}
		  }
		}`,
		Variables: map[string]interface{}{},
	}
	request, err := graphQLRequest.ToRequest(ctx, d.settings.URL)
	if err != nil {
		return nil, err
	}

	resp, err := d.httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	graphQLResponse, responseParseError := graphql.ParseGraphQLResponse(resp.Body)
	if responseParseError != nil {
		if resp.StatusCode == 200 {
			return &backend.CheckHealthResult{
				Status:  backend.HealthStatusError,
				Message: "Successful status code, but could not parse GraphQL response",
			}, nil
		}
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Could not parse GraphQL response! Got status: " + resp.Status,
		}, nil
	}
	if len(graphQLResponse.Errors) > 0 {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "GraphQL response contained errors! HTTP status: " + resp.Status,
		}, nil
	}
	if resp.StatusCode != 200 {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Something went wrong: " + resp.Status,
		}, nil
	}
	_, schemaExists := graphQLResponse.Data["__schema"]
	if !schemaExists {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Unexpected GraphQL response!",
		}, nil
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Success",
	}, nil
}
