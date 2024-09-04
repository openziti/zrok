package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/admin"
	"github.com/sirupsen/logrus"
)

type grantsHandler struct{}

func newGrantsHandler() *grantsHandler {
	return &grantsHandler{}
}

func (h *grantsHandler) Handle(params admin.GrantsParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		logrus.Errorf("invalid admin principal")
		return admin.NewGrantsUnauthorized()
	}
	return admin.NewGrantsOK()
}
