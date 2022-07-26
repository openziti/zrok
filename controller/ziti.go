package controller

import (
	"crypto/x509"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/edge/rest_util"
)

func edgeClient() (*rest_management_api_client.ZitiEdgeManagement, error) {
	ctrlAddress := "https://linux:1280"
	caCerts, err := rest_util.GetControllerWellKnownCas(ctrlAddress)
	if err != nil {
		return nil, err
	}
	caPool := x509.NewCertPool()
	for _, ca := range caCerts {
		caPool.AddCert(ca)
	}
	return rest_util.NewEdgeManagementClientWithUpdb("admin", "admin", ctrlAddress, caPool)
}
