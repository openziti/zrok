package automation

import (
	"crypto/x509"

	"github.com/openziti/edge-api/rest_management_api_client"
	"github.com/openziti/edge-api/rest_util"
)

type Config struct {
	ApiEndpoint string
	Username    string
	Password    string `df:",secret"`
}

type Client struct {
	edge *rest_management_api_client.ZitiEdgeManagement
}

func NewClient(cfg *Config) (*Client, error) {
	caCerts, err := rest_util.GetControllerWellKnownCas(cfg.ApiEndpoint)
	if err != nil {
		return nil, err
	}
	caPool := x509.NewCertPool()
	for _, ca := range caCerts {
		caPool.AddCert(ca)
	}
	edge, err := rest_util.NewEdgeManagementClientWithUpdb(cfg.Username, cfg.Password, cfg.ApiEndpoint, caPool)
	if err != nil {
		return nil, err
	}
	return &Client{edge: edge}, nil
}

func (c *Client) Edge() *rest_management_api_client.ZitiEdgeManagement {
	return c.edge
}
