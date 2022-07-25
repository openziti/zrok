package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/rest_model"
	"github.com/openziti-test-kitchen/zrok/rest_zrok_server/operations/identity"
	"github.com/sirupsen/logrus"
)

func enableHandler(params identity.EnableParams) middleware.Responder {
	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return middleware.Error(500, err.Error())
	}
	a, err := str.FindAccountWithToken(params.Body.Token, tx)
	if err != nil {
		logrus.Errorf("error finding account: %v", err)
		return middleware.Error(500, err.Error())
	}
	if a == nil {
		logrus.Errorf("account not found: %v", err)
		return middleware.Error(404, err.Error())
	}
	logrus.Infof("found account '%v'", a.Username)

	return identity.NewEnableCreated().WithPayload(&rest_model.EnableResponse{
		Identity: a.Username,
	})
}
