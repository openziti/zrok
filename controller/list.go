package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/metadata"
	"github.com/sirupsen/logrus"
)

func listIdentitiesHandler(params metadata.ListIdentitiesParams, principal *rest_model_zrok.Principal) middleware.Responder {
	logrus.Infof("principal: %v", principal.Username)
	return metadata.NewListIdentitiesOK()
}
