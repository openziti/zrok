package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/admin"
)

type createAccountHandler struct{}

func newCreateAccountHandler() *createAccountHandler {
	return &createAccountHandler{}
}

func (h *createAccountHandler) Handle(params admin.CreateAccountParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		dl.Error("invalid admin principal")
		return admin.NewCreateAccountUnauthorized()
	}

	token, err := CreateToken()
	if err != nil {
		dl.Errorf("error creating token: %v", err)
		return admin.NewCreateAccountInternalServerError()
	}
	hpwd, err := HashPassword(params.Body.Password)
	if err != nil {
		dl.Errorf("error hashing password: %v", err)
		return admin.NewCreateAccountInternalServerError()
	}

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return admin.NewCreateAccountInternalServerError()
	}
	defer trx.Rollback()

	a := &store.Account{
		Email:    params.Body.Email,
		Salt:     hpwd.Salt,
		Password: hpwd.Password,
		Token:    token,
	}
	if _, err := str.CreateAccount(a, trx); err != nil {
		dl.Errorf("error creating account: %v", err)
		return admin.NewCreateAccountInternalServerError()
	}
	if err := trx.Commit(); err != nil {
		dl.Errorf("error committing transaction: %v", err)
	}

	dl.Infof("administratively created account '%v'", params.Body.Email)

	return admin.NewCreateAccountCreated().WithPayload(&admin.CreateAccountCreatedBody{AccountToken: token})
}
