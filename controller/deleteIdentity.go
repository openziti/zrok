package controller

import (
	"context"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/edge-api/rest_management_api_client/identity"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/admin"
	"github.com/sirupsen/logrus"
)

type deleteIdentityHandler struct{}

func newDeleteIdentityHandler() *deleteIdentityHandler {
	return &deleteIdentityHandler{}
}

func (h *deleteIdentityHandler) Handle(params admin.DeleteIdentityParams, principal *rest_model_zrok.Principal) middleware.Responder {
	identityZId := params.Body.ZID

	if !principal.Admin {
		logrus.Errorf("invalid admin principal")
		return admin.NewDeleteIdentityUnauthorized()
	}

	edge, err := zrokEdgeSdk.Client(cfg.Ziti)
	if err != nil {
		logrus.Errorf("error getting edge client: %v", err)
		return admin.NewDeleteIdentityInternalServerError()
	}

	req := &identity.DeleteIdentityParams{
		ID:      identityZId,
		Context: context.Background(),
	}
	req.SetTimeout(30 * time.Second)
	if _, err := edge.Identity.DeleteIdentity(req, nil); err != nil {
		logrus.Errorf("error deleting identity '%v': %v", identityZId, err)
		return admin.NewDeleteIdentityInternalServerError()
	}

	return admin.NewDeleteIdentityOK()
}
