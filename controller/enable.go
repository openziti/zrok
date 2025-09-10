package controller

import (
	"bytes"
	"encoding/json"

	"github.com/go-openapi/runtime/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/controller/automation"
	"github.com/openziti/zrok/controller/store"
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

	automationCfg := &automation.Config{
		ApiEndpoint: cfg.Ziti.ApiEndpoint,
		Username:    cfg.Ziti.Username,
		Password:    cfg.Ziti.Password,
	}

	za, err := automation.NewZitiAutomation(automationCfg)
	if err != nil {
		logrus.Errorf("error connecting to ziti edge management api for user '%v': %v", principal.Email, err)
		return environment.NewEnableInternalServerError()
	}

	uniqueToken, err := createShareToken()
	if err != nil {
		logrus.Errorf("error creating unique identity token for user '%v': %v", principal.Email, err)
		return environment.NewEnableInternalServerError()
	}

	envZId, zitiCfg, err := h.createEnvironmentIdentity(za, uniqueToken, principal.Email, params.Body.Description)
	if err != nil {
		logrus.Errorf("error creating environment identity for user '%v': %v", principal.Email, err)
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
	logrus.Infof("created environment for '%v', with ziti identity '%v', and database id '%v'", principal.Email, envZId, envId)

	resp := environment.NewEnableCreated().WithPayload(&environment.EnableCreatedBody{Identity: envZId})

	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	enc.SetEscapeHTML(false)
	err = enc.Encode(&zitiCfg)
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

func (h *enableHandler) createEnvironmentIdentity(za *automation.ZitiAutomation, uniqueToken, accountEmail, envDescription string) (string, interface{}, error) {
	return za.CreateEnvironmentIdentity(uniqueToken, accountEmail, envDescription)
}
