package controller

import (
	"bytes"
	"encoding/json"
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/controller/store"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/environment"
	"github.com/sirupsen/logrus"
)

type enableHandler struct {
}

func newEnableHandler() *enableHandler {
	return &enableHandler{}
}

func (h *enableHandler) Handle(params environment.EnableParams, principal *rest_model_zrok.Principal) middleware.Responder {
	// start transaction early; if it fails, don't bother creating ziti resources
	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return environment.NewEnableInternalServerError()
	}

	client, err := edgeClient()
	if err != nil {
		logrus.Errorf("error getting edge client: %v", err)
		return environment.NewEnableInternalServerError()
	}
	ident, err := createIdentity(principal.Email, client)
	if err != nil {
		logrus.Error(err)
		return environment.NewEnableInternalServerError()
	}
	cfg, err := enrollIdentity(ident.Payload.Data.ID, client)
	if err != nil {
		logrus.Error(err)
		return environment.NewEnableInternalServerError()
	}
	if err := createEdgeRouterPolicy(ident.Payload.Data.ID, client); err != nil {
		logrus.Error(err)
		return environment.NewEnableInternalServerError()
	}
	envId, err := str.CreateEnvironment(int(principal.ID), &store.Environment{
		Description: params.Body.Description,
		Host:        params.Body.Host,
		Address:     realRemoteAddress(params.HTTPRequest),
		ZId:         ident.Payload.Data.ID,
	}, tx)
	if err != nil {
		logrus.Errorf("error storing created identity: %v", err)
		_ = tx.Rollback()
		return environment.NewEnableInternalServerError()
	}
	if err := tx.Commit(); err != nil {
		logrus.Errorf("error committing: %v", err)
		return environment.NewEnableInternalServerError()
	}
	logrus.Infof("created environment for '%v', with ziti identity '%v', and database id '%v'", principal.Email, ident.Payload.Data.ID, envId)

	resp := environment.NewEnableCreated().WithPayload(&rest_model_zrok.EnableResponse{
		Identity: ident.Payload.Data.ID,
	})

	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	enc.SetEscapeHTML(false)
	err = enc.Encode(&cfg)
	if err != nil {
		panic(err)
	}
	resp.Payload.Cfg = out.String()

	return resp
}
