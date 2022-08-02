package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/identity"
	"github.com/sirupsen/logrus"
)

func loginHandler(params identity.LoginParams) middleware.Responder {
	if params.Body == nil || params.Body.Email == "" || params.Body.Password == "" {
		logrus.Errorf("missing email or password")
		return identity.NewLoginUnauthorized()
	}

	logrus.Infof("received login request for email '%v'", params.Body.Email)

	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return identity.NewLoginUnauthorized()
	}
	defer func() { _ = tx.Rollback() }()
	a, err := str.FindAccountWithUsername(params.Body.Email, tx)
	if err != nil {
		logrus.Errorf("error finding account '%v': %v", params.Body.Email, err)
		return identity.NewLoginUnauthorized()
	}
	if a.Password != hashPassword(params.Body.Password) {
		logrus.Errorf("password mismatch for account '%v'", params.Body.Email)
		return identity.NewLoginUnauthorized()
	}

	return identity.NewLoginOK().WithPayload(rest_model_zrok.LoginResponse(a.Token))
}
