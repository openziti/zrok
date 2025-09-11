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
	edge                      *rest_management_api_client.ZitiEdgeManagement
	Identity                  *IdentityManager
	Service                   *ServiceManager
	Config                    *ConfigManager
	ConfigType                *ConfigTypeManager
	EdgeRouterPolicy          *EdgeRouterPolicyManager
	ServiceEdgeRouterPolicy   *ServiceEdgeRouterPolicyManager
	ServicePolicy             *ServicePolicyManager
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
	
	client := &Client{edge: edge}
	client.Identity = NewIdentityManager(client)
	client.Service = NewServiceManager(client)
	client.Config = NewConfigManager(client)
	client.ConfigType = NewConfigTypeManager(client)
	client.EdgeRouterPolicy = NewEdgeRouterPolicyManager(client)
	client.ServiceEdgeRouterPolicy = NewServiceEdgeRouterPolicyManager(client)
	client.ServicePolicy = NewServicePolicyManager(client)
	
	return client, nil
}

func (c *Client) Edge() *rest_management_api_client.ZitiEdgeManagement {
	return c.edge
}
