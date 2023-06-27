package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/account"
	"github.com/sirupsen/logrus"
)

type verifyHandler struct {
}

func newVerifyHandler() *verifyHandler {
	return &verifyHandler{}
}

func (self *verifyHandler) Handle(params account.VerifyParams) middleware.Responder {
	if params.Body != nil {
		logrus.Debugf("received verify request for token '%v'", params.Body.Token)
		tx, err := str.Begin()
		if err != nil {
			logrus.Errorf("error starting transaction: %v", err)
			return account.NewVerifyInternalServerError()
		}
		defer func() { _ = tx.Rollback() }()

		ar, err := str.FindAccountRequestWithToken(params.Body.Token, tx)
		if err != nil {
			logrus.Errorf("error finding account request with token '%v': %v", params.Body.Token, err)
			return account.NewVerifyNotFound()
		}

		return account.NewVerifyOK().WithPayload(&rest_model_zrok.VerifyResponse{Email: ar.Email})
	}
	logrus.Error("empty verification request")
	return account.NewVerifyInternalServerError()
}
