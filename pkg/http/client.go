package http

import (
	"net/http"
	"net/http/httputil"
)

type Client struct {
	http.Client
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) PostRequest(req *http.Request) (*http.Response, error) {
	if _, err := httputil.DumpRequestOut(req, true); err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
