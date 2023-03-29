package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/share"
	"github.com/sirupsen/logrus"
)

type accessHandler struct{}

func newAccessHandler() *accessHandler {
	return &accessHandler{}
}

func (h *accessHandler) Handle(params share.AccessParams, principal *rest_model_zrok.Principal) middleware.Responder {
	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction for user '%v': %v", principal.Email, err)
		return share.NewAccessInternalServerError()
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
			return share.NewAccessUnauthorized()
		}
	} else {
		logrus.Errorf("error finding environments for account '%v'", principal.Email)
		return share.NewAccessNotFound()
	}

	shrToken := params.Body.ShrToken
	shr, err := str.FindShareWithToken(shrToken, tx)
	if err != nil {
		logrus.Errorf("error finding share")
		return share.NewAccessNotFound()
	}
	if shr == nil {
		logrus.Errorf("unable to find share '%v' for user '%v'", shrToken, principal.Email)
		return share.NewAccessNotFound()
	}

	feToken, err := createToken()
	if err != nil {
		logrus.Error(err)
		return share.NewAccessInternalServerError()
	}

	if _, err := str.CreateFrontend(envId, &store.Frontend{Token: feToken, ZId: envZId}, tx); err != nil {
		logrus.Errorf("error creating frontend record for user '%v': %v", principal.Email, err)
		return share.NewAccessInternalServerError()
	}

	edge, err := zrokEdgeSdk.Client(cfg.Ziti)
	if err != nil {
		logrus.Error(err)
		return share.NewAccessInternalServerError()
	}
	addlTags := map[string]interface{}{
		"zrokEnvironmentZId": envZId,
		"zrokFrontendToken":  feToken,
		"zrokShareToken":     shrToken,
	}
	if err := zrokEdgeSdk.CreateServicePolicyDial(envZId+"-"+shr.ZId+"-dial", shr.ZId, []string{envZId}, addlTags, edge); err != nil {
		logrus.Errorf("unable to create dial policy for user '%v': %v", principal.Email, err)
		return share.NewAccessInternalServerError()
	}

	if err := tx.Commit(); err != nil {
		logrus.Errorf("error committing frontend record: %v", err)
		return share.NewAccessInternalServerError()
	}

	return share.NewAccessCreated().WithPayload(&rest_model_zrok.AccessResponse{FrontendToken: feToken})
}
