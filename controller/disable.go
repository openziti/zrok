package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/identity"
	"github.com/sirupsen/logrus"
)

type disableHandler struct {
	cfg *Config
}

func newDisableHandler(cfg *Config) *disableHandler {
	return &disableHandler{cfg: cfg}
}

func (self *disableHandler) Handle(params identity.DisableParams, principal *rest_model_zrok.Principal) middleware.Responder {
	_, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return identity.NewDisableInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	return identity.NewDisableOK()
}
