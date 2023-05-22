package zrokEdgeSdk

import (
	"crypto/x509"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/edge/rest_util"
)

type Config struct {
	ApiEndpoint string
	Username    string
	Password    string `cf:"+secret"`
}

func Client(cfg *Config) (*rest_management_api_client.ZitiEdgeManagement, error) {
	caCerts, err := rest_util.GetControllerWellKnownCas(cfg.ApiEndpoint)
	if err != nil {
		return nil, err
	}
	caPool := x509.NewCertPool()
	for _, ca := range caCerts {
		caPool.AddCert(ca)
	}
	return rest_util.NewEdgeManagementClientWithUpdb(cfg.Username, cfg.Password, cfg.ApiEndpoint, caPool)
}
