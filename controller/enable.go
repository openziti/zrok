package controller

import (
	"bytes"
	"encoding/json"
	"github.com/go-openapi/runtime/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/environment"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type enableHandler struct{}

func newEnableHandler() *enableHandler {
	return &enableHandler{}
}

func (h *enableHandler) Handle(params environment.EnableParams, principal *rest_model_zrok.Principal) middleware.Responder {
	// start transaction early; if it fails, don't bother creating ziti resources
	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction for user '%v': %v", principal.Email, err)
		return environment.NewEnableInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	if err := h.checkLimits(principal, trx); err != nil {
		logrus.Errorf("limits error for user '%v': %v", principal.Email, err)
		return environment.NewEnableUnauthorized()
	}

	client, err := zrokEdgeSdk.Client(cfg.Ziti)
	if err != nil {
		logrus.Errorf("error getting edge client for user '%v': %v", principal.Email, err)
		return environment.NewEnableInternalServerError()
	}

	uniqueToken, err := createShareToken()
	if err != nil {
		logrus.Errorf("error creating unique identity token for user '%v': %v", principal.Email, err)
		return environment.NewEnableInternalServerError()
	}

	ident, err := zrokEdgeSdk.CreateEnvironmentIdentity(uniqueToken, principal.Email, params.Body.Description, client)
	if err != nil {
		logrus.Errorf("error creating environment identity for user '%v': %v", principal.Email, err)
		return environment.NewEnableInternalServerError()
	}

	envZId := ident.Payload.Data.ID
	cfg, err := zrokEdgeSdk.EnrollIdentity(envZId, client)
	if err != nil {
		logrus.Errorf("error enrolling environment identity for user '%v': %v", principal.Email, err)
		return environment.NewEnableInternalServerError()
	}

	if err := zrokEdgeSdk.CreateEdgeRouterPolicy(envZId, envZId, client); err != nil {
		logrus.Errorf("error creating edge router policy for user '%v': %v", principal.Email, err)
		return environment.NewEnableInternalServerError()
	}

	envId, err := str.CreateEnvironment(int(principal.ID), &store.Environment{
		Description: params.Body.Description,
		Host:        params.Body.Host,
		Address:     realRemoteAddress(params.HTTPRequest),
		ZId:         envZId,
	}, trx)
	if err != nil {
		logrus.Errorf("error storing created identity for user '%v': %v", principal.Email, err)
		_ = trx.Rollback()
		return environment.NewEnableInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		logrus.Errorf("error committing for user '%v': %v", principal.Email, err)
		return environment.NewEnableInternalServerError()
	}
	logrus.Infof("created environment for '%v', with ziti identity '%v', and database id '%v'", principal.Email, ident.Payload.Data.ID, envId)

	resp := environment.NewEnableCreated().WithPayload(&environment.EnableCreatedBody{Identity: envZId})

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

func (h *enableHandler) checkLimits(principal *rest_model_zrok.Principal, trx *sqlx.Tx) error {
	if !principal.Limitless {
		if limitsAgent != nil {
			ok, err := limitsAgent.CanCreateEnvironment(int(principal.ID), trx)
			if err != nil {
				return errors.Wrapf(err, "error checking environment limits for '%v'", principal.Email)
			}
			if !ok {
				return errors.Errorf("environment limit check failed for '%v'", principal.Email)
			}
		}
	}
	return nil
}
