package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/admin"
	"github.com/sirupsen/logrus"
)

type deleteOrganizationHandler struct{}

func newDeleteOrganizationHandler() *deleteOrganizationHandler {
	return &deleteOrganizationHandler{}
}

func (h *deleteOrganizationHandler) Handle(params admin.DeleteOrganizationParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		logrus.Errorf("invalid admin principal")
		return admin.NewDeleteOrganizationUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return admin.NewDeleteOrganizationInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	org, err := str.FindOrganizationByToken(params.Body.OrganizationToken, trx)
	if err != nil {
		logrus.Errorf("error finding organization by token: %v", err)
		return admin.NewDeleteOrganizationNotFound()
	}

	err = str.DeleteOrganization(org.Id, trx)
	if err != nil {
		logrus.Errorf("error deleting organization: %v", err)
		return admin.NewDeleteOrganizationInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		logrus.Errorf("error committing transaction: %v", err)
		return admin.NewDeleteOrganizationInternalServerError()
	}

	return admin.NewDeleteOrganizationOK()
}
