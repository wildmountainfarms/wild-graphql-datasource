package graphql

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/wildmountainfarms/wild-graphql-datasource/pkg/util/jsonnode"
	"io"
	"net/http"
)

type Request struct {
	Query string `json:"query"`
	// A map of variable names to the value of that variable. Allowed value types are strings, numeric types, and booleans
	Variables     map[string]interface{} `json:"variables,omitempty"`
	OperationName string                 `json:"operationName,omitempty"`
}

func (request *Request) ToBody() ([]byte, error) {
	return json.Marshal(request)
}
func (request *Request) ToRequest(ctx context.Context, url string) (*http.Request, error) {
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
	httpReq.Header.Add("Accept", "application/json")

	return httpReq, nil
}

type Response struct {
	// We use a jsonnode.Object here because it maintains the order of the keys in a JSON object
	Data   *jsonnode.Object `json:"data"`
	Errors []Error          `json:"errors"`
}

type Error struct {
	Message    string                 `json:"message"`
	Locations  []ErrorLocation        `json:"locations,omitempty"`
	Path       []interface{}          `json:"path,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

type ErrorLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

func ParseGraphQLResponse(body io.ReadCloser) (*Response, error) {
	bodyAsBytes, err := io.ReadAll(body)
	if err != nil {
		log.DefaultLogger.Error("We don't expect this!")
		return nil, err
	}
	var graphQLResponse Response
	err = json.Unmarshal(bodyAsBytes, &graphQLResponse)
	if err != nil {
		log.DefaultLogger.Error("Error while parsing GraphQL response to graphql.Response")
		return nil, err
	}
	return &graphQLResponse, nil
}
