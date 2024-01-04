package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/httpclient"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"io"
	"net/http"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/data"
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
	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res, err := d.query(ctx, req.PluginContext, q)
		if err != nil {
			// If an error is returned from the query, we assume that it is not a recoverable error.
			//   We can consider changing this in the future
			return nil, err
		}

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = *res // we don't expect the result to be nil if err is not nil
	}

	return response, nil
}

// queryModel represents data sent from the frontend to perform a query
type queryModel struct {
	query string
}

type graphQLRequest struct {
	Query string `json:"query"`
	// A map of variable names to the value of that variable. Allowed value types are strings, numeric types, and booleans
	Variables map[string]interface{} `json:"variables,omitempty"`
}

func (request *graphQLRequest) toBody() ([]byte, error) {
	return json.Marshal(request)
}

func (d *Datasource) query(_ context.Context, pCtx backend.PluginContext, query backend.DataQuery) (*backend.DataResponse, error) {
	var response backend.DataResponse

	// Unmarshal the JSON into our queryModel.
	var qm queryModel

	err := json.Unmarshal(query.JSON, &qm)
	if err != nil {
		// We don't
		return nil, err
	}
	// Although the frontend has access to global variable substitution (https://grafana.com/docs/grafana/latest/dashboards/variables/add-template-variables/#global-variables)
	//   the backend does not.
	//   Because of this, it's beneficial to encourage users to write queries that rely as little on the frontend as possible.
	//   This allows us to support alerting later.
	//   These variable names that we are "hard coding" should be as similar to possible as those global variables that are available
	//   Forum post here: https://community.grafana.com/t/how-to-use-template-variables-in-your-data-source/63250#backend-data-sources-3
	//   More information here: https://grafana.com/docs/grafana/latest/dashboards/variables/

	//requestJson := graphQLRequest{
	request := graphQLRequest{
		Query: qm.query,
		Variables: map[string]interface{}{
			"from":        query.TimeRange.From.UnixMilli(),
			"to":          query.TimeRange.To.UnixMilli(),
			"interval_ms": query.Interval.Milliseconds(),
		},
	}
	_, err = request.toBody()
	if err != nil {
		// We don't actually expect this to happen, so we can return an "unrecoverable" error
		return nil, err
		//return backend.ErrDataResponse(backend.StatusInternal, fmt.Sprintf("Internal request.toBody() error: %v", err.Error()))
	}

	// TODO follow this tutorial: https://www.thepolyglotdeveloper.com/2020/02/interacting-with-a-graphql-api-with-golang/

	// create data frame response.
	// For an overview on data frames and how grafana handles them:
	// https://grafana.com/developers/plugin-tools/introduction/data-frames
	frame := data.NewFrame("response")

	// add fields.
	frame.Fields = append(frame.Fields,
		data.NewField("time", nil, []time.Time{query.TimeRange.From, query.TimeRange.To}),
		data.NewField("values", nil, []int64{10, 20}),
	)

	// add the frames to the response.
	response.Frames = append(response.Frames, frame)

	return &response, nil
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	// test command to do the same thing:
	//   curl -X POST -H "Content-Type: application/json" -d '{"query":"{\n\t\t  __schema{\n\t\t\tqueryType{name}\n\t\t  }\n\t\t}"}' https://swapi-graphql.netlify.app/.netlify/functions/index
	request := graphQLRequest{
		Query: `{
		  __schema{
			queryType{name}
		  }
		}`,
		Variables: map[string]interface{}{},
	}
	body, err := request.toBody()
	if err != nil {
		return nil, nil
	}
	log.DefaultLogger.Info(fmt.Sprintf("Body: %s", body))

	httpReq, err := http.NewRequestWithContext(ctx, "POST", d.settings.URL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	// if we don't add this header, we get an error of "Must provide query string"
	httpReq.Header.Add("Content-Type", "application/json")

	resp, err := d.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		responseBody, bodyReadErr := io.ReadAll(resp.Body)
		if bodyReadErr != nil {
			log.DefaultLogger.Error("We don't expect this!")
			return nil, bodyReadErr
		}
		log.DefaultLogger.Info(fmt.Sprintf("Response: %s", responseBody))

		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Something went wrong: " + resp.Status,
		}, nil
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Success",
	}, nil
}
