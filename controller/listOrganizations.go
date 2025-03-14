package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/admin"
	"github.com/sirupsen/logrus"
)

type listOrganizationsHandler struct{}

func newListOrganizationsHandler() *listOrganizationsHandler {
	return &listOrganizationsHandler{}
}

func (h *listOrganizationsHandler) Handle(_ admin.ListOrganizationsParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		logrus.Error("invalid admin principal")
		return admin.NewListOrganizationsUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return admin.NewListOrganizationsInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	orgs, err := str.FindOrganizations(trx)
	if err != nil {
		logrus.Errorf("error finding organizations: %v", err)
		return admin.NewListOrganizationsInternalServerError()
	}

	var out []*admin.ListOrganizationsOKBodyOrganizationsItems0
	for _, org := range orgs {
		out = append(out, &admin.ListOrganizationsOKBodyOrganizationsItems0{Description: org.Description, OrganizationToken: org.Token})
	}
	return admin.NewListOrganizationsOK().WithPayload(&admin.ListOrganizationsOKBody{Organizations: out})
}
