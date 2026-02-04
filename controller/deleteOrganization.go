package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/admin"
)

type deleteOrganizationHandler struct{}

func newDeleteOrganizationHandler() *deleteOrganizationHandler {
	return &deleteOrganizationHandler{}
}

func (h *deleteOrganizationHandler) Handle(params admin.DeleteOrganizationParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		dl.Errorf("invalid admin principal")
		return admin.NewDeleteOrganizationUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return admin.NewDeleteOrganizationInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	org, err := str.FindOrganizationByToken(params.Body.OrganizationToken, trx)
	if err != nil {
		dl.Errorf("error finding organization by token: %v", err)
		return admin.NewDeleteOrganizationNotFound()
	}

	err = str.DeleteOrganization(org.Id, trx)
	if err != nil {
		dl.Errorf("error deleting organization: %v", err)
		return admin.NewDeleteOrganizationInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		dl.Errorf("error committing transaction: %v", err)
		return admin.NewDeleteOrganizationInternalServerError()
	}

	return admin.NewDeleteOrganizationOK()
}
