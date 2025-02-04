package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/rest_server_zrok/operations/account"
	"github.com/sirupsen/logrus"
)

type verifyHandler struct{}

func newVerifyHandler() *verifyHandler {
	return &verifyHandler{}
}

func (h *verifyHandler) Handle(params account.VerifyParams) middleware.Responder {
	if params.Body.RegisterToken != "" {
		logrus.Debugf("received verify request for registration token '%v'", params.Body.RegisterToken)
		tx, err := str.Begin()
		if err != nil {
			logrus.Errorf("error starting transaction: %v", err)
			return account.NewVerifyInternalServerError()
		}
		defer func() { _ = tx.Rollback() }()

		ar, err := str.FindAccountRequestWithToken(params.Body.RegisterToken, tx)
		if err != nil {
			logrus.Errorf("error finding account request with registration token '%v': %v", params.Body.RegisterToken, err)
			return account.NewVerifyNotFound()
		}

		return account.NewVerifyOK().WithPayload(&account.VerifyOKBody{Email: ar.Email})
	}
	logrus.Error("empty verification request")
	return account.NewVerifyInternalServerError()
}
