package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/service"
	"github.com/sirupsen/logrus"
)

type updateShareHandler struct{}

func newUpdateShareHandler() *updateShareHandler {
	return &updateShareHandler{}
}

func (h *updateShareHandler) Handle(params service.UpdateShareParams, principal *rest_model_zrok.Principal) middleware.Responder {
	svcToken := params.Body.ServiceToken
	backendProxyEndpoint := params.Body.BackendProxyEndpoint

	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return service.NewUpdateShareInternalServerError()
	}
	defer func() { _ = tx.Rollback() }()

	sshr, err := str.FindShareWithToken(svcToken, tx)
	if err != nil {
		logrus.Errorf("service '%v' not found: %v", svcToken, err)
		return service.NewUpdateShareNotFound()
	}

	senvs, err := str.FindEnvironmentsForAccount(int(principal.ID), tx)
	if err != nil {
		logrus.Errorf("error finding environments for account '%v': %v", principal.Email, err)
		return service.NewUpdateShareInternalServerError()
	}

	envFound := false
	for _, senv := range senvs {
		if senv.Id == sshr.Id {
			envFound = true
			break
		}
	}
	if !envFound {
		logrus.Errorf("environment not found for service '%v'", svcToken)
		return service.NewUpdateShareNotFound()
	}

	sshr.BackendProxyEndpoint = &backendProxyEndpoint
	if err := str.UpdateShare(sshr, tx); err != nil {
		logrus.Errorf("error updating service '%v': %v", svcToken, err)
		return service.NewUpdateShareInternalServerError()
	}

	if err := tx.Commit(); err != nil {
		logrus.Errorf("error committing transaction for service '%v' update: %v", svcToken, err)
		return service.NewUpdateShareInternalServerError()
	}

	return service.NewUpdateShareOK()
}
