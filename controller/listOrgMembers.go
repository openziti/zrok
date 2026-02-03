package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/metadata"
)

type listOrgMembersHandler struct{}

func newListOrgMembersHandler() *listOrgMembersHandler {
	return &listOrgMembersHandler{}
}

func (h *listOrgMembersHandler) Handle(params metadata.ListOrgMembersParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return metadata.NewListOrgMembersInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	org, err := str.FindOrganizationByToken(params.OrganizationToken, trx)
	if err != nil {
		dl.Errorf("error finding organization by token: %v", err)
		return metadata.NewListOrgMembersNotFound()
	}

	admin, err := str.IsAccountAdminOfOrganization(int(principal.ID), org.Id, trx)
	if err != nil {
		dl.Errorf("error checking account '%v' admin: %v", principal.Email, err)
		return metadata.NewListOrgMembersNotFound()
	}
	if !admin {
		dl.Errorf("requesting account '%v' is not admin of organization '%v'", principal.Email, org.Token)
		return metadata.NewOrgAccountOverviewNotFound()
	}

	members, err := str.FindAccountsForOrganization(org.Id, trx)
	if err != nil {
		dl.Errorf("error finding accounts for organization '%v': %v", org.Token, err)
		return metadata.NewListOrgMembersInternalServerError()
	}

	var out []*metadata.ListOrgMembersOKBodyMembersItems0
	for _, member := range members {
		out = append(out, &metadata.ListOrgMembersOKBodyMembersItems0{Email: member.Email, Admin: member.Admin})
	}
	return metadata.NewListOrgMembersOK().WithPayload(&metadata.ListOrgMembersOKBody{Members: out})
}
