package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/rest_server_zrok/operations/account"
	"github.com/sirupsen/logrus"
)

func loginHandler(params account.LoginParams) middleware.Responder {
	if params.Body.Email == "" || params.Body.Password == "" {
		logrus.Errorf("missing email or password")
		return account.NewLoginUnauthorized()
	}

	logrus.Infof("received login request for email '%v'", params.Body.Email)

	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return account.NewLoginUnauthorized()
	}
	defer func() { _ = tx.Rollback() }()
	a, err := str.FindAccountWithEmail(params.Body.Email, tx)
	if err != nil {
		logrus.Errorf("error finding account '%v': %v", params.Body.Email, err)
		return account.NewLoginUnauthorized()
	}
	hpwd, err := rehashPassword(params.Body.Password, a.Salt)
	if err != nil {
		logrus.Errorf("error hashing password for '%v': %v", params.Body.Email, err)
		return account.NewLoginUnauthorized()
	}
	if a.Password != hpwd.Password {
		logrus.Errorf("password mismatch for account '%v'", params.Body.Email)
		return account.NewLoginUnauthorized()
	}

	return account.NewLoginOK().WithPayload(a.Token)
}
