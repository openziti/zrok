package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/controller/store"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/identity"
	"github.com/sirupsen/logrus"
)

type createAccountHandler struct {
	cfg *Config
}

func newCreateAccountHandler(cfg *Config) *createAccountHandler {
	return &createAccountHandler{cfg: cfg}
}

func (self *createAccountHandler) Handle(params identity.CreateAccountParams) middleware.Responder {
	logrus.Infof("received account request for email '%v'", params.Body.Email)
	if params.Body == nil || params.Body.Email == "" {
		logrus.Errorf("missing email")
		return identity.NewCreateAccountBadRequest().WithPayload("missing email")
	}
	token := createToken()
	if err := sendVerificationEmail(params.Body.Email, token, self.cfg); err != nil {
		logrus.Error(err)
		return identity.NewCreateAccountInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	ar := &store.AccountRequest{
		Token:         token,
		Email:         params.Body.Email,
		SourceAddress: params.HTTPRequest.RemoteAddr,
	}
	tx, err := str.Begin()
	if err != nil {
		logrus.Error(err)
		return identity.NewCreateAccountInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := str.CreateAccountRequest(ar, tx); err != nil {
		logrus.Error(err)
		return identity.NewCreateAccountInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	if err := tx.Commit(); err != nil {
		logrus.Error(err)
		return identity.NewCreateAccountInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	return identity.NewCreateAccountCreated()
}
