package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/controller/store"
	"github.com/openziti-test-kitchen/zrok/controller/zrokEdgeSdk"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/service"
	"github.com/sirupsen/logrus"
)

type accessHandler struct{}

func newAccessHandler() *accessHandler {
	return &accessHandler{}
}

func (h *accessHandler) Handle(params service.AccessParams, principal *rest_model_zrok.Principal) middleware.Responder {
	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return service.NewAccessInternalServerError()
	}
	defer func() { _ = tx.Rollback() }()

	envZId := params.Body.EnvZID
	envId := 0
	if envs, err := str.FindEnvironmentsForAccount(int(principal.ID), tx); err == nil {
		found := false
		for _, env := range envs {
			if env.ZId == envZId {
				logrus.Debugf("found identity '%v' for user '%v'", envZId, principal.Email)
				envId = env.Id
				found = true
				break
			}
		}
		if !found {
			logrus.Errorf("environment '%v' not found for user '%v'", envZId, principal.Email)
			return service.NewAccessUnauthorized()
		}
	} else {
		logrus.Errorf("error finding environments for account '%v'", principal.Email)
		return service.NewAccessNotFound()
	}

	svcToken := params.Body.SvcToken
	ssvc, err := str.FindServiceWithToken(svcToken, tx)
	if err != nil {
		logrus.Errorf("error finding service")
		return service.NewAccessNotFound()
	}
	if ssvc == nil {
		logrus.Errorf("unable to find service '%v' for user '%v'", svcToken, principal.Email)
		return service.NewAccessNotFound()
	}

	feToken, err := createToken()
	if err != nil {
		logrus.Error(err)
		return service.NewAccessInternalServerError()
	}

	if _, err := str.CreateFrontend(envId, &store.Frontend{Token: feToken, ZId: envZId}, tx); err != nil {
		logrus.Errorf("error creating frontend record: %v", err)
		return service.NewAccessInternalServerError()
	}

	edge, err := edgeClient()
	if err != nil {
		logrus.Error(err)
		return service.NewAccessInternalServerError()
	}
	addlTags := map[string]interface{}{
		"zrokEnvironmentZId": envZId,
		"zrokFrontendToken":  feToken,
		"zrokServiceToken":   svcToken,
	}
	if err := zrokEdgeSdk.CreateServicePolicyDial(envZId+"-"+ssvc.ZId+"-dial", ssvc.ZId, []string{envZId}, addlTags, edge); err != nil {
		logrus.Errorf("unable to create dial policy: %v", err)
		return service.NewAccessInternalServerError()
	}

	if err := tx.Commit(); err != nil {
		logrus.Errorf("error committing frontend record: %v", err)
		return service.NewAccessInternalServerError()
	}

	return service.NewAccessCreated().WithPayload(&rest_model_zrok.AccessResponse{FrontendToken: feToken})
}
