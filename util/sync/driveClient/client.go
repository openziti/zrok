package driveClient

import "net/http"

type Client struct {
	client *http.Client
}

func NewHttpClient() *Client {
	return &Client{&http.Client{}}
}

func (c *Client) Connect() error {
	return nil
}

func (c *Client) options(uri string) (*http.Response, error) {
	return c.request("OPTIONS", uri)
}

func (c *Client) request(method, uri string) (resp *http.Response, err error) {
	req, err := http.NewRequest(method, uri, nil)
	if err != nil {
		return nil, err
	}

	if resp, err = c.client.Do(req); err != nil {
		return resp, err
	}

	return resp, err
}
