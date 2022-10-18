package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/controller/store"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/identity"
	"github.com/sirupsen/logrus"
)

type registerHandler struct{}

func newRegisterHandler() *registerHandler {
	return &registerHandler{}
}
func (self *registerHandler) Handle(params identity.RegisterParams) middleware.Responder {
	if params.Body == nil || params.Body.Token == "" || params.Body.Password == "" {
		logrus.Error("missing token or password")
		return identity.NewRegisterNotFound()
	}
	logrus.Infof("received register request for token '%v'", params.Body.Token)

	tx, err := str.Begin()
	if err != nil {
		logrus.Error(err)
		return identity.NewRegisterInternalServerError()
	}
	defer func() { _ = tx.Rollback() }()

	ar, err := str.FindAccountRequestWithToken(params.Body.Token, tx)
	if err != nil {
		logrus.Error(err)
		return identity.NewRegisterNotFound()
	}

	token, err := createToken()
	if err != nil {
		logrus.Error(err)
		return identity.NewRegisterInternalServerError()
	}
	a := &store.Account{
		Email:    ar.Email,
		Password: hashPassword(params.Body.Password),
		Token:    token,
	}
	if _, err := str.CreateAccount(a, tx); err != nil {
		logrus.Error(err)
		return identity.NewRegisterInternalServerError()
	}

	if err := str.DeleteAccountRequest(ar.Id, tx); err != nil {
		logrus.Error(err)
		return identity.NewRegisterInternalServerError()
	}

	if err := tx.Commit(); err != nil {
		logrus.Error(err)
		return identity.NewCreateAccountInternalServerError()
	}

	logrus.Infof("created account '%v' with token '%v'", a.Email, a.Token)

	return identity.NewRegisterOK().WithPayload(&rest_model_zrok.RegisterResponse{Token: a.Token})
}
