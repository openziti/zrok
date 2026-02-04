package controller

import (
	"fmt"

	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/automation"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/admin"
)

type deleteIdentityHandler struct{}

func newDeleteIdentityHandler() *deleteIdentityHandler {
	return &deleteIdentityHandler{}
}

func (h *deleteIdentityHandler) Handle(params admin.DeleteIdentityParams, principal *rest_model_zrok.Principal) middleware.Responder {
	identityZId := params.Body.ZID

	if !principal.Admin {
		dl.Errorf("invalid admin principal")
		return admin.NewDeleteIdentityUnauthorized()
	}

	ziti, err := automation.NewZitiAutomation(cfg.Ziti)
	if err != nil {
		dl.Errorf("error getting automation client: %v", err)
		return admin.NewDeleteIdentityInternalServerError()
	}

	// delete edge router policy for the identity
	erpFilter := fmt.Sprintf("name=\"%v\"", identityZId)
	if err := ziti.EdgeRouterPolicies.DeleteWithFilter(erpFilter); err != nil {
		dl.Errorf("error deleting edge router policy: %v", err)
		return admin.NewDeleteIdentityInternalServerError()
	}

	// delete the identity
	if err := ziti.Identities.Delete(identityZId); err != nil {
		dl.Errorf("error deleting identity '%v': %v", identityZId, err)
		return admin.NewDeleteIdentityInternalServerError()
	}

	return admin.NewDeleteIdentityOK()
}
