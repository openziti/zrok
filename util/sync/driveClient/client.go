package driveClient

import "net/http"

type Client struct {
	client *http.Client
}

func NewHttpClient(uri string) *Client {
	return &Client{&http.Client{}}
}

func (c *Client) Connect() error {
	return nil
}
