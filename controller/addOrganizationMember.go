package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/admin"
)

type addOrganizationMemberHandler struct{}

func newAddOrganizationMemberHandler() *addOrganizationMemberHandler {
	return &addOrganizationMemberHandler{}
}

func (h *addOrganizationMemberHandler) Handle(params admin.AddOrganizationMemberParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		dl.Error("invalid admin principal")
		return admin.NewAddOrganizationMemberUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return admin.NewAddOrganizationMemberInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	acct, err := str.FindAccountWithEmail(params.Body.Email, trx)
	if err != nil {
		dl.Errorf("error finding account with email address '%v': %v", params.Body.Email, err)
		return admin.NewAddOrganizationMemberNotFound()
	}

	org, err := str.FindOrganizationByToken(params.Body.OrganizationToken, trx)
	if err != nil {
		dl.Errorf("error finding organization '%v': %v", params.Body.OrganizationToken, err)
		return admin.NewAddOrganizationMemberNotFound()
	}

	if err := str.AddAccountToOrganization(acct.Id, org.Id, params.Body.Admin, trx); err != nil {
		dl.Errorf("error adding account '%v' to organization '%v': %v", acct.Email, org.Token, err)
		return admin.NewAddOrganizationMemberInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		dl.Errorf("error committing transaction: %v", err)
		return admin.NewAddOrganizationMemberInternalServerError()
	}

	return admin.NewAddOrganizationMemberCreated()
}
