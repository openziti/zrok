package controller

import (
	"fmt"

	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/admin"
	"github.com/sirupsen/logrus"
)

type deleteFrontendGrantHandler struct{}

func newDeleteFrontendGrantHandler() *deleteFrontendGrantHandler {
	return &deleteFrontendGrantHandler{}
}

func (h *deleteFrontendGrantHandler) Handle(params admin.DeleteFrontendGrantParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		logrus.Error("invalid admin principal")
		return admin.NewDeleteFrontendGrantUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return admin.NewDeleteFrontendGrantInternalServerError()
	}
	defer trx.Rollback()

	fe, err := str.FindFrontendWithToken(params.Body.FrontendToken, trx)
	if err != nil {
		logrus.Errorf("error finding frontend with token '%v': %v", params.Body.FrontendToken, err)
		return admin.NewDeleteFrontendGrantNotFound().WithPayload(rest_model_zrok.ErrorMessage(fmt.Sprintf("frontend token '%v' not found", params.Body.FrontendToken)))
	}

	acct, err := str.FindAccountWithEmail(params.Body.Email, trx)
	if err != nil {
		logrus.Errorf("error finding account with email '%v': %v", params.Body.Email, err)
		return admin.NewDeleteFrontendGrantNotFound().WithPayload(rest_model_zrok.ErrorMessage(fmt.Sprintf("account '%v' not found", params.Body.Email)))
	}

	if granted, err := str.IsFrontendGrantedToAccount(fe.Id, acct.Id, trx); err != nil {
		logrus.Errorf("error checking frontend grant for account '%v' and frontend '%v': %v", acct.Email, fe.Token, err)
		return admin.NewDeleteFrontendGrantInternalServerError()

	} else if granted {
		if err := str.DeleteFrontendGrant(fe.Id, acct.Id, trx); err != nil {
			logrus.Errorf("error deleting frontend ('%v') grant for '%v': %v", fe.Token, acct.Email, err)
			return admin.NewDeleteFrontendGrantInternalServerError()
		}
		logrus.Infof("deleted '%v' access to frontend '%v'", acct.Email, fe.Token)

	} else {
		logrus.Infof("account '%v' not granted access to frontend '%v'", acct.Email, fe.Token)
	}

	return admin.NewDeleteFrontendGrantOK()
}
