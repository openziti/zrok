package controller

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	errors2 "github.com/go-openapi/errors"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/edge/rest_util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func ZrokAuthenticate(token string) (*rest_model_zrok.Principal, error) {
	logrus.Infof("authenticating")
	tx, err := str.Begin()
	if err != nil {
		return nil, err
	}
	if a, err := str.FindAccountWithToken(token, tx); err == nil {
		principal := rest_model_zrok.Principal{
			ID:       int64(a.Id),
			Token:    a.Token,
			Username: a.Username,
		}
		return &principal, nil
	} else {
		return nil, errors2.New(401, "invalid api key")
	}
}

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

func generateApiToken() (string, error) {
	bytes := make([]byte, 64)
	if _, err := rand.Read(bytes); err != nil {
		return "", errors.Wrap(err, "error generating random api token")
	}
	return hex.EncodeToString(bytes), nil
}

func randomId() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", errors.Wrap(err, "error generating random identity id")
	}
	return hex.EncodeToString(bytes), nil
}
