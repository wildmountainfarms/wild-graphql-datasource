package mock

import (
	"bytes"
	"io"
	"net/http"
)

type Client struct {
	resp []byte
}

func NewClient(resp []byte) *Client {
	return &Client{resp: resp}
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		Body:       io.NopCloser(bytes.NewReader(c.resp)),
		StatusCode: http.StatusOK,
	}, nil
}
