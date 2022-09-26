package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/identity"
	"github.com/sirupsen/logrus"
)

type verifyHandler struct {
	cfg *Config
}

func newVerifyHandler(cfg *Config) *verifyHandler {
	return &verifyHandler{cfg: cfg}
}

func (self *verifyHandler) Handle(params identity.VerifyParams) middleware.Responder {
	if params.Body != nil {
		logrus.Debugf("received verify request for token '%v'", params.Body.Token)
		tx, err := str.Begin()
		if err != nil {
			logrus.Errorf("error starting transaction: %v", err)
			return identity.NewVerifyInternalServerError()
		}
		defer func() { _ = tx.Rollback() }()

		ar, err := str.FindAccountRequestWithToken(params.Body.Token, tx)
		if err != nil {
			logrus.Errorf("error finding account with token '%v': %v", params.Body.Token, err)
			return identity.NewVerifyNotFound()
		}

		return identity.NewVerifyOK().WithPayload(&rest_model_zrok.VerifyResponse{Email: ar.Email})
	} else {
		logrus.Error("empty verification request")
		return identity.NewVerifyInternalServerError().WithPayload(rest_model_zrok.ErrorMessage("empty verification request"))
	}
}
