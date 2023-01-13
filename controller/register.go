package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/account"
	"github.com/sirupsen/logrus"
)

type registerHandler struct{}

func newRegisterHandler() *registerHandler {
	return &registerHandler{}
}
func (self *registerHandler) Handle(params account.RegisterParams) middleware.Responder {
	if params.Body == nil || params.Body.Token == "" || params.Body.Password == "" {
		logrus.Error("missing token or password")
		return account.NewRegisterNotFound()
	}
	logrus.Infof("received register request for token '%v'", params.Body.Token)

	tx, err := str.Begin()
	if err != nil {
		logrus.Error(err)
		return account.NewRegisterInternalServerError()
	}
	defer func() { _ = tx.Rollback() }()

	ar, err := str.FindAccountRequestWithToken(params.Body.Token, tx)
	if err != nil {
		logrus.Error(err)
		return account.NewRegisterNotFound()
	}

	token, err := createToken()
	if err != nil {
		logrus.Error(err)
		return account.NewRegisterInternalServerError()
	}
	a := &store.Account{
		Email:    ar.Email,
		Password: hashPassword(params.Body.Password),
		Token:    token,
	}
	if _, err := str.CreateAccount(a, tx); err != nil {
		logrus.Error(err)
		return account.NewRegisterInternalServerError()
	}

	if err := str.DeleteAccountRequest(ar.Id, tx); err != nil {
		logrus.Error(err)
		return account.NewRegisterInternalServerError()
	}

	if err := tx.Commit(); err != nil {
		logrus.Error(err)
		return account.NewRegisterInternalServerError()
	}

	logrus.Infof("created account '%v' with token '%v'", a.Email, a.Token)

	return account.NewRegisterOK().WithPayload(&rest_model_zrok.RegisterResponse{Token: a.Token})
}
