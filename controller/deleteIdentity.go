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

	// Try to get identity by ID first, then by name if not found
	identity, err := ziti.Identities.GetByID(identityZId)
	if err != nil {
		if ziti.IsNotFound(err) {
			// Try by name
			identity, err = ziti.Identities.GetByName(identityZId)
			if err != nil {
				if ziti.IsNotFound(err) {
					dl.Warnf("identity '%v' not found by ID or name, treating as already deleted", identityZId)
					return admin.NewDeleteIdentityOK()
				}
				dl.Errorf("error looking up identity '%v': %v", identityZId, err)
				return admin.NewDeleteIdentityInternalServerError()
			}
		} else {
			dl.Errorf("error looking up identity '%v': %v", identityZId, err)
			return admin.NewDeleteIdentityInternalServerError()
		}
	}

	// Use the actual ID for deletion
	actualId := *identity.ID

	// delete edge router policy for the identity
	erpFilter := fmt.Sprintf("name=\"%v\"", actualId)
	if err := ziti.EdgeRouterPolicies.DeleteWithFilter(erpFilter); err != nil {
		dl.Errorf("error deleting edge router policy: %v", err)
		return admin.NewDeleteIdentityInternalServerError()
	}

	// delete the identity
	if err := ziti.Identities.Delete(actualId); err != nil {
		// treat 404 as success (idempotent delete)
		if ziti.IsNotFound(err) {
			dl.Warnf("identity '%v' not found in Ziti, treating as already deleted", actualId)
			return admin.NewDeleteIdentityOK()
		}
		dl.Errorf("error deleting identity '%v': %v", actualId, err)
		return admin.NewDeleteIdentityInternalServerError()
	}

	dl.Infof("deleted identity '%v' (id: %v)", *identity.Name, actualId)
	return admin.NewDeleteIdentityOK()
}
