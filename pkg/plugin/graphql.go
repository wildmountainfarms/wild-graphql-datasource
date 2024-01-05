package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"io"
	"net/http"
)

// NOTE: We may consider using this library in the future instead: https://github.com/machinebox/graphql

type GraphQLRequest struct {
	Query string `json:"query"`
	// A map of variable names to the value of that variable. Allowed value types are strings, numeric types, and booleans
	Variables     map[string]interface{} `json:"variables,omitempty"`
	OperationName string                 `json:"operationName,omitempty"`
}

func (request *GraphQLRequest) ToBody() ([]byte, error) {
	return json.Marshal(request)
}
func (request *GraphQLRequest) ToRequest(ctx context.Context, url string) (*http.Request, error) {
	body, err := request.ToBody()
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	// if we don't add this header, we get an error of "Must provide query string"
	httpReq.Header.Add("Content-Type", "application/json")

	return httpReq, nil
}

type GraphQLResponse struct {
	Data   map[string]interface{} `json:"data"`
	Errors []GraphQLError         `json:"errors"`
}

type GraphQLError struct {
	Message    string                 `json:"message"`
	Locations  []GraphQLErrorLocation `json:"locations,omitempty"`
	Path       []interface{}          `json:"path,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

type GraphQLErrorLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

func ParseGraphQLResponse(body io.ReadCloser) (*GraphQLResponse, error) {
	bodyAsBytes, err := io.ReadAll(body)
	if err != nil {
		log.DefaultLogger.Error("We don't expect this!")
		return nil, err
	}
	var graphQLResponse GraphQLResponse
	err = json.Unmarshal(bodyAsBytes, &graphQLResponse)
	if err != nil {
		return nil, err
	}
	return &graphQLResponse, nil
}
