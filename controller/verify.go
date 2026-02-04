package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/account"
)

type verifyHandler struct{}

func newVerifyHandler() *verifyHandler {
	return &verifyHandler{}
}

func (h *verifyHandler) Handle(params account.VerifyParams) middleware.Responder {
	if params.Body.RegisterToken != "" {
		dl.Debugf("received verify request for registration token '%v'", params.Body.RegisterToken)
		trx, err := str.Begin()
		if err != nil {
			dl.Errorf("error starting transaction: %v", err)
			return account.NewVerifyInternalServerError()
		}
		defer func() { _ = trx.Rollback() }()

		ar, err := str.FindAccountRequestWithToken(params.Body.RegisterToken, trx)
		if err != nil {
			dl.Errorf("error finding account request with registration token '%v': %v", params.Body.RegisterToken, err)
			return account.NewVerifyNotFound()
		}

		return account.NewVerifyOK().WithPayload(&account.VerifyOKBody{Email: ar.Email})
	}
	dl.Error("empty verification request")
	return account.NewVerifyInternalServerError()
}
