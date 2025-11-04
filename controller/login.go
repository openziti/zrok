package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/rest_server_zrok/operations/account"
)

func loginHandler(params account.LoginParams) middleware.Responder {
	if params.Body.Email == "" || params.Body.Password == "" {
		dl.Errorf("missing email or password")
		return account.NewLoginUnauthorized()
	}

	dl.Infof("received login request for email '%v'", params.Body.Email)

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return account.NewLoginUnauthorized()
	}
	defer func() { _ = trx.Rollback() }()
	a, err := str.FindAccountWithEmail(params.Body.Email, trx)
	if err != nil {
		dl.Errorf("error finding account '%v': %v", params.Body.Email, err)
		return account.NewLoginUnauthorized()
	}
	hpwd, err := rehashPassword(params.Body.Password, a.Salt)
	if err != nil {
		dl.Errorf("error hashing password for '%v': %v", params.Body.Email, err)
		return account.NewLoginUnauthorized()
	}
	if a.Password != hpwd.Password {
		dl.Errorf("password mismatch for account '%v'", params.Body.Email)
		return account.NewLoginUnauthorized()
	}

	return account.NewLoginOK().WithPayload(a.Token)
}
