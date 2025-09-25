package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/admin"
)

type listOrganizationMembersHandler struct{}

func newListOrganizationMembersHandler() *listOrganizationMembersHandler {
	return &listOrganizationMembersHandler{}
}

func (h *listOrganizationMembersHandler) Handle(params admin.ListOrganizationMembersParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		dl.Error("invalid admin principal")
		return admin.NewListOrganizationMembersUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return admin.NewListOrganizationMembersInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	org, err := str.FindOrganizationByToken(params.Body.OrganizationToken, trx)
	if err != nil {
		dl.Errorf("error finding organization by token: %v", err)
		return admin.NewListOrganizationMembersNotFound()
	}
	if org == nil {
		dl.Errorf("organization '%v' not found", params.Body.OrganizationToken)
		return admin.NewListOrganizationMembersNotFound()
	}

	oms, err := str.FindAccountsForOrganization(org.Id, trx)
	if err != nil {
		dl.Errorf("error finding accounts for organization: %v", err)
		return admin.NewListOrganizationMembersInternalServerError()
	}

	var members []*admin.ListOrganizationMembersOKBodyMembersItems0
	for _, om := range oms {
		members = append(members, &admin.ListOrganizationMembersOKBodyMembersItems0{Email: om.Email, Admin: om.Admin})
	}
	return admin.NewListOrganizationMembersOK().WithPayload(&admin.ListOrganizationMembersOKBody{Members: members})
}
