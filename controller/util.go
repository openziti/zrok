package controller

import (
	"crypto/rand"
	"crypto/sha512"
	"crypto/x509"
	"encoding/hex"
	errors2 "github.com/go-openapi/errors"
	"github.com/lithammer/shortuuid/v4"
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

func createToken() string {
	return shortuuid.New()
}

func createServiceName() (string, error) {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", errors.Wrap(err, "error generating service name")
	}
	return hex.EncodeToString(bytes), nil
}

func hashPassword(raw string) string {
	hash := sha512.New()
	hash.Write([]byte(raw))
	return hex.EncodeToString(hash.Sum(nil))
}
