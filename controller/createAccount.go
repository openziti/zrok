package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/admin"
	"github.com/sirupsen/logrus"
)

type createAccountHandler struct{}

func newCreateAccountHandler() *createAccountHandler {
	return &createAccountHandler{}
}

func (h *createAccountHandler) Handle(params admin.CreateAccountParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		logrus.Errorf("invalid admin principal")
		return admin.NewCreateAccountUnauthorized()
	}

	token, err := CreateToken()
	if err != nil {
		logrus.Errorf("error creating token: %v", err)
		return admin.NewCreateAccountInternalServerError()
	}
	hpwd, err := HashPassword(params.Body.Password)
	if err != nil {
		logrus.Errorf("error hashing password: %v", err)
		return admin.NewCreateAccountInternalServerError()
	}

	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return admin.NewCreateAccountInternalServerError()
	}
	defer func() {
		_ = trx.Rollback()
	}()
	a := &store.Account{
		Email:    params.Body.Email,
		Salt:     hpwd.Salt,
		Password: hpwd.Password,
		Token:    token,
	}
	if _, err := str.CreateAccount(a, trx); err != nil {
		logrus.Errorf("error creating account: %v", err)
		return admin.NewCreateAccountInternalServerError()
	}
	if err := trx.Commit(); err != nil {
		logrus.Errorf("error committing transaction: %v", err)
	}

	logrus.Infof("administratively created account '%v'", params.Body.Email)

	return admin.NewCreateAccountCreated().WithPayload(&admin.CreateAccountCreatedBody{AccountToken: token})
}
