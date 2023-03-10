package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/share"
	"github.com/sirupsen/logrus"
)

type updateShareHandler struct{}

func newUpdateShareHandler() *updateShareHandler {
	return &updateShareHandler{}
}

func (h *updateShareHandler) Handle(params share.UpdateShareParams, principal *rest_model_zrok.Principal) middleware.Responder {
	shrToken := params.Body.ShrToken
	backendProxyEndpoint := params.Body.BackendProxyEndpoint

	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return share.NewUpdateShareInternalServerError()
	}
	defer func() { _ = tx.Rollback() }()

	sshr, err := str.FindShareWithToken(shrToken, tx)
	if err != nil {
		logrus.Errorf("share '%v' not found: %v", shrToken, err)
		return share.NewUpdateShareNotFound()
	}

	senvs, err := str.FindEnvironmentsForAccount(int(principal.ID), tx)
	if err != nil {
		logrus.Errorf("error finding environments for account '%v': %v", principal.Email, err)
		return share.NewUpdateShareInternalServerError()
	}

	envFound := false
	for _, senv := range senvs {
		if !senv.Deleted && senv.Id == sshr.Id {
			envFound = true
			break
		}
	}
	if !envFound {
		logrus.Errorf("environment not found for share '%v'", shrToken)
		return share.NewUpdateShareNotFound()
	}

	sshr.BackendProxyEndpoint = &backendProxyEndpoint
	if err := str.UpdateShare(sshr, tx); err != nil {
		logrus.Errorf("error updating share '%v': %v", shrToken, err)
		return share.NewUpdateShareInternalServerError()
	}

	if err := tx.Commit(); err != nil {
		logrus.Errorf("error committing transaction for share '%v' update: %v", shrToken, err)
		return share.NewUpdateShareInternalServerError()
	}

	return share.NewUpdateShareOK()
}
