package controller

import (
	"crypto/sha512"
	"encoding/hex"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/controller/store"
	"github.com/openziti-test-kitchen/zrok/rest_model"
	"github.com/openziti-test-kitchen/zrok/rest_zrok_server"
	"github.com/openziti-test-kitchen/zrok/rest_zrok_server/operations"
	"github.com/openziti-test-kitchen/zrok/rest_zrok_server/operations/identity"
	"github.com/openziti-test-kitchen/zrok/rest_zrok_server/operations/metadata"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var str *store.Store

func Run(cfg *Config) error {
	if v, err := store.Open(cfg.Store); err == nil {
		str = v
	} else {
		return errors.Wrap(err, "error opening store")
	}

	swaggerSpec, err := loads.Embedded(rest_zrok_server.SwaggerJSON, rest_zrok_server.FlatSwaggerJSON)
	if err != nil {
		return errors.Wrap(err, "error loading embedded swagger spec")
	}

	api := operations.NewZrokAPI(swaggerSpec)
	api.MetadataVersionHandler = metadata.VersionHandlerFunc(versionHandler)
	api.IdentityCreateAccountHandler = identity.CreateAccountHandlerFunc(createAccountHandler)

	server := rest_zrok_server.NewServer(api)
	defer func() { _ = server.Shutdown() }()
	server.Host = cfg.Host
	server.Port = cfg.Port
	server.ConfigureAPI()
	if err := server.Serve(); err != nil {
		return errors.Wrap(err, "api server error")
	}
	return nil
}

func versionHandler(_ metadata.VersionParams) middleware.Responder {
	return metadata.NewGetOK().WithPayload(&rest_model.Version{Version: "v0.0.0; sk3tch"})
}

func createAccountHandler(params identity.CreateAccountParams) middleware.Responder {
	logrus.Infof("received account request for username '%v'", params.Body.Username)
	if params.Body == nil || params.Body.Username == "" || params.Body.Password == "" {
		return middleware.Error(500, errors.Errorf("invalid username or password"))
	}

	token, err := generateApiToken()
	if err != nil {
		logrus.Errorf("error generating api token: %v", err)
		return middleware.Error(500, err.Error())
	}

	a := &store.Account{
		Username: params.Body.Username,
		Password: hashPassword(params.Body.Password),
		Token:    token,
	}
	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return middleware.Error(500, err.Error())
	}
	id, err := str.CreateAccount(a, tx)
	if err != nil {
		logrus.Errorf("error creating account: %v", err)
		_ = tx.Rollback()
		return middleware.Error(400, err.Error())
	}
	if err := tx.Commit(); err != nil {
		logrus.Errorf("error comitting: %v", err)
	}

	logrus.Infof("account created with id = '%v'", id)
	return identity.NewCreateAccountCreated().WithPayload(&rest_model.AccountResponse{
		APIToken: token,
	})
}

func hashPassword(raw string) string {
	hash := sha512.New()
	hash.Write([]byte(raw))
	return hex.EncodeToString(hash.Sum(nil))
}
