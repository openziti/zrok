package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/service"
)

type shareHandler struct{}

func newShareHandler() *shareHandler {
	return &shareHandler{}
}

func (h *shareHandler) Handle(params service.ShareParams, principal *rest_model_zrok.Principal) middleware.Responder {
	switch params.Body.ShareMode {
	case "public":
		return newSharePublicHandler().Handle(params, principal)
	default:
		return service.NewShareInternalServerError()
	}
}
