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
	if params.Body == nil || params.Body.Email == "" {
		logrus.Errorf("missing email")
		return identity.NewCreateAccountBadRequest().WithPayload("missing email")
	}
	logrus.Infof("received account request for email '%v'", params.Body.Email)

	token := createToken()
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

	if _, err := str.FindAccountWithEmail(params.Body.Email, tx); err == nil {
		logrus.Errorf("found account for '%v', cannot process account request", params.Body.Email)
		return identity.NewCreateAccountBadRequest()
	} else {
		logrus.Infof("no account found for '%v': %v", params.Body.Email, err)
	}

	if oldAr, err := str.FindAccountRequestWithEmail(params.Body.Email, tx); err == nil {
		logrus.Warnf("found previous account request for '%v', removing", params.Body.Email)
		if err := str.DeleteAccountRequest(oldAr.Id, tx); err != nil {
			logrus.Errorf("error deleteing previous account request for '%v': %v", params.Body.Email, err)
			return identity.NewCreateAccountInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
		}
	} else {
		logrus.Warnf("error finding previous account request for '%v': %v", params.Body.Email, err)
	}

	if _, err := str.CreateAccountRequest(ar, tx); err != nil {
		logrus.Error(err)
		return identity.NewCreateAccountInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	if err := tx.Commit(); err != nil {
		logrus.Error(err)
		return identity.NewCreateAccountInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}

	if err := sendVerificationEmail(params.Body.Email, token, self.cfg); err != nil {
		logrus.Error(err)
		return identity.NewCreateAccountInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}

	return identity.NewCreateAccountCreated()
}
