package controller

import (
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/share"
	"github.com/sirupsen/logrus"
)

type unaccessHandler struct{}

func newUnaccessHandler() *unaccessHandler {
	return &unaccessHandler{}
}

func (h *unaccessHandler) Handle(params share.UnaccessParams, principal *rest_model_zrok.Principal) middleware.Responder {
	feToken := params.Body.FrontendToken
	shrToken := params.Body.ShareToken
	envZId := params.Body.EnvZID
	logrus.Infof("processing unaccess request for frontend '%v' (share '%v', environment '%v')", feToken, shrToken, envZId)

	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return share.NewUnaccessInternalServerError()
	}
	defer func() { _ = tx.Rollback() }()

	edge, err := zrokEdgeSdk.Client(cfg.Ziti)
	if err != nil {
		logrus.Error(err)
		return share.NewUnaccessInternalServerError()
	}

	var senv *store.Environment
	if envs, err := str.FindEnvironmentsForAccount(int(principal.ID), tx); err == nil {
		for _, env := range envs {
			if env.ZId == envZId {
				senv = env
				break
			}
		}
		if senv == nil {
			logrus.Errorf("environment with id '%v' not found for '%v", envZId, principal.Email)
			return share.NewUnaccessUnauthorized()
		}
	} else {
		logrus.Errorf("error finding environments for account '%v': %v", principal.Email, err)
		return share.NewUnaccessUnauthorized()
	}

	sfe, err := str.FindFrontendWithToken(feToken, tx)
	if err != nil {
		logrus.Errorf("error finding frontend for '%v': %v", principal.Email, err)
		return share.NewUnaccessInternalServerError()
	}

	if sfe == nil || (sfe.EnvironmentId != nil && *sfe.EnvironmentId != senv.Id) {
		logrus.Errorf("frontend named '%v' not found", feToken)
		return share.NewUnaccessInternalServerError()
	}

	if err := str.DeleteFrontend(sfe.Id, tx); err != nil {
		logrus.Errorf("error deleting frontend named '%v': %v", feToken, err)
		return share.NewUnaccessNotFound()
	}

	if err := zrokEdgeSdk.DeleteServicePolicies(envZId, fmt.Sprintf("tags.zrokShareToken=\"%v\" and tags.zrokFrontendToken=\"%v\" and type=1", shrToken, feToken), edge); err != nil {
		logrus.Errorf("error removing access to '%v' for '%v': %v", shrToken, envZId, err)
		return share.NewUnaccessInternalServerError()
	}

	if err := tx.Commit(); err != nil {
		logrus.Errorf("error committing frontend '%v' delete: %v", feToken, err)
		return share.NewUnaccessInternalServerError()
	}

	return share.NewUnaccessOK()
}
