package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/share"
	"github.com/sirupsen/logrus"
)

type updateAccessHandler struct{}

func newUpdateAccessHandler() *updateAccessHandler {
	return &updateAccessHandler{}
}

func (h *updateAccessHandler) Handle(params share.UpdateAccessParams, principal *rest_model_zrok.Principal) middleware.Responder {
	frontendToken := params.Body.FrontendToken
	bindAddress := params.Body.BindAddress
	desc := params.Body.Description

	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return share.NewUpdateAccessInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	fe, err := str.FindFrontendWithToken(frontendToken, trx)
	if err != nil {
		logrus.Errorf("error finding frontend with token '%v': %v", frontendToken, err)
		return share.NewUpdateAccessNotFound()
	}

	envs, err := str.FindEnvironmentsForAccount(int(principal.ID), trx)
	if err != nil {
		logrus.Errorf("error finding environments for account '%v': %v", principal.Email, err)
	}

	envMatched := false
	for _, env := range envs {
		if fe.EnvironmentId != nil && env.Id == *fe.EnvironmentId {
			envMatched = true
			break
		}
	}
	if !envMatched {
		logrus.Errorf("account '%v' does not own frontend '%v'", principal.Email, frontendToken)
		return share.NewUpdateAccessNotFound()
	}

	if desc != "" {
		fe.Description = &desc
	} else {
		fe.Description = nil
	}
	if bindAddress != "" {
		fe.BindAddress = &bindAddress
	} else {
		fe.BindAddress = nil
	}
	if err := str.UpdateFrontend(fe, trx); err != nil {
		logrus.Errorf("error updating frontend '%v': %v", frontendToken, err)
		return share.NewUpdateAccessInternalServerError()
	}
	if err := trx.Commit(); err != nil {
		logrus.Errorf("error committing transaction for frontend '%v': %v", frontendToken, err)
		return share.NewUpdateAccessInternalServerError()
	}
	return share.NewUpdateAccessOK()
}
