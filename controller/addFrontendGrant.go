package controller

import (
	"fmt"

	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/admin"
	"github.com/sirupsen/logrus"
)

type addFrontendGrantHandler struct{}

func newAddFrontendGrantHandler() *addFrontendGrantHandler {
	return &addFrontendGrantHandler{}
}

func (h *addFrontendGrantHandler) Handle(params admin.AddFrontendGrantParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		logrus.Error("invalid admin principal")
		return admin.NewAddFrontendGrantUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return admin.NewAddFrontendGrantInternalServerError()
	}
	defer trx.Rollback()

	fe, err := str.FindFrontendWithToken(params.Body.FrontendToken, trx)
	if err != nil {
		logrus.Errorf("error finding frontend with token '%v': %v", params.Body.FrontendToken, err)
		return admin.NewAddFrontendGrantNotFound().WithPayload(rest_model_zrok.ErrorMessage(fmt.Sprintf("frontend token '%v' not found", params.Body.FrontendToken)))
	}

	acct, err := str.FindAccountWithEmail(params.Body.Email, trx)
	if err != nil {
		logrus.Errorf("error finding account with email '%v': %v", params.Body.Email, err)
		return admin.NewAddFrontendGrantNotFound().WithPayload(rest_model_zrok.ErrorMessage(fmt.Sprintf("account '%v' not found", params.Body.Email)))
	}

	if granted, err := str.IsFrontendGrantedToAccount(fe.Id, acct.Id, trx); err != nil {
		logrus.Errorf("error checking frontend grant for account '%v' and frontend '%v': %v", acct.Email, fe.Token, err)
		return admin.NewAddFrontendGrantInternalServerError()

	} else if !granted {
		if _, err := str.CreateFrontendGrant(fe.Id, acct.Id, trx); err != nil {
			logrus.Errorf("error creating frontend ('%v') grant for '%v': %v", fe.Token, acct.Email, err)
			return admin.NewAddFrontendGrantInternalServerError()
		}
		logrus.Infof("granted '%v' access to frontend '%v'", acct.Email, fe.Token)
		
	} else {
		logrus.Infof("account '%v' already granted access to frontend '%v'", acct.Email, fe.Token)
	}

	return admin.NewAddFrontendGrantOK()
}
