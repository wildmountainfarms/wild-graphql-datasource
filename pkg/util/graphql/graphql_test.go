package graphql

import (
	"os"
	"testing"
)

func TestParsing(t *testing.T) {
	fi, err := os.Open("testdata/example_graphql_response.json")
	if err != nil {
		t.Fatal(err)
		return
	}
	_, err = ParseGraphQLResponse(fi)
	if err != nil {
		t.Fatal(err)
		return
	}

}
