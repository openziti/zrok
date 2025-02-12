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

	shrToken := params.Body.ShareToken
	shr, err := str.FindShareWithToken(shrToken, trx)
	if err != nil {
		logrus.Errorf("error finding share with token '%v': %v", shrToken, err)
		return share.NewAccessNotFound()
	}
	if shr == nil {
		logrus.Errorf("unable to find share '%v' for user '%v'", shrToken, principal.Email)
		return share.NewAccessNotFound()
	}

	if shr.PermissionMode == store.ClosedPermissionMode {
		shrEnv, err := str.GetEnvironment(shr.EnvironmentId, trx)
		if err != nil {
			logrus.Errorf("error getting environment for share '%v': %v", shrToken, err)
			return share.NewAccessInternalServerError()
		}

		if err := h.checkAccessGrants(shr, *shrEnv.AccountId, principal, trx); err != nil {
			logrus.Errorf("closed permission mode for '%v' fails for '%v': %v", shr.Token, principal.Email, err)
			return share.NewAccessUnauthorized()
		}
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

	if _, err := str.CreateFrontend(envId, &store.Frontend{PrivateShareId: &shr.Id, Token: feToken, ZId: envZId, PermissionMode: store.ClosedPermissionMode}, trx); err != nil {
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

	return share.NewAccessCreated().WithPayload(&share.AccessCreatedBody{
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

func (h *accessHandler) checkAccessGrants(shr *store.Share, ownerAccountId int, principal *rest_model_zrok.Principal, trx *sqlx.Tx) error {
	if int(principal.ID) == ownerAccountId {
		logrus.Infof("accessing own share '%v' for '%v'", shr.Token, principal.Email)
		return nil
	}
	count, err := str.CheckAccessGrantForShareAndAccount(shr.Id, int(principal.ID), trx)
	if err != nil {
		logrus.Infof("error checking access grants for '%v': %v", shr.Token, err)
		return err
	}
	if count > 0 {
		logrus.Infof("found '%d' grants for '%v'", count, principal.Email)
		return nil
	}
	return errors.Errorf("access denied for '%v' accessing '%v'", principal.Email, shr.Token)
}
