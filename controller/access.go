package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/share"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type accessHandler struct{}

func newAccessHandler() *accessHandler {
	return &accessHandler{}
}

func (h *accessHandler) Handle(params share.AccessParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction for user '%v': %v", principal.Email, err)
		return share.NewAccessInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	envZId := params.Body.EnvZID
	envId := 0
	if envs, err := str.FindEnvironmentsForAccount(int(principal.ID), trx); err == nil {
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
	shr, err := str.FindShareWithToken(shrToken, trx)
	if err != nil {
		logrus.Errorf("error finding share")
		return share.NewAccessNotFound()
	}
	if shr == nil {
		logrus.Errorf("unable to find share '%v' for user '%v'", shrToken, principal.Email)
		return share.NewAccessNotFound()
	}

	if err := h.checkLimits(shr, trx); err != nil {
		logrus.Errorf("cannot access limited share for '%v': %v", principal.Email, err)
		return share.NewAccessNotFound()
	}

	feToken, err := CreateToken()
	if err != nil {
		logrus.Error(err)
		return share.NewAccessInternalServerError()
	}

	if _, err := str.CreateFrontend(envId, &store.Frontend{PrivateShareId: &shr.Id, Token: feToken, ZId: envZId}, trx); err != nil {
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
	if err := zrokEdgeSdk.CreateServicePolicyDial(feToken+"-"+envZId+"-"+shr.ZId+"-dial", shr.ZId, []string{envZId}, addlTags, edge); err != nil {
		logrus.Errorf("unable to create dial policy for user '%v': %v", principal.Email, err)
		return share.NewAccessInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		logrus.Errorf("error committing frontend record: %v", err)
		return share.NewAccessInternalServerError()
	}

	return share.NewAccessCreated().WithPayload(&rest_model_zrok.AccessResponse{
		FrontendToken: feToken,
		BackendMode:   shr.BackendMode,
	})
}

func (h *accessHandler) checkLimits(shr *store.Share, trx *sqlx.Tx) error {
	if limitsAgent != nil {
		ok, err := limitsAgent.CanAccessShare(shr.Id, trx)
		if err != nil {
			return errors.Wrapf(err, "error checking share limits for '%v'", shr.Token)
		}
		if !ok {
			return errors.Errorf("share limit check failed for '%v'", shr.Token)
		}
	}
	return nil
}
