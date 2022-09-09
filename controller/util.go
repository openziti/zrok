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
)

func ZrokAuthenticate(token string) (*rest_model_zrok.Principal, error) {
	tx, err := str.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if a, err := str.FindAccountWithToken(token, tx); err == nil {
		principal := rest_model_zrok.Principal{
			ID:    int64(a.Id),
			Token: a.Token,
			Email: a.Email,
		}
		return &principal, nil
	} else {
		return nil, errors2.New(401, "invalid api key")
	}
}

func edgeClient(cfg *ZitiConfig) (*rest_management_api_client.ZitiEdgeManagement, error) {
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
