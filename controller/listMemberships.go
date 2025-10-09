package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/metadata"
)

type listMembershipsHandler struct{}

func newListMembershipsHandler() *listMembershipsHandler {
	return &listMembershipsHandler{}
}

func (h *listMembershipsHandler) Handle(_ metadata.ListMembershipsParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return metadata.NewListMembershipsInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	oms, err := str.FindOrganizationsForAccount(int(principal.ID), trx)
	if err != nil {
		dl.Errorf("error finding organizations for account '%v': %v", principal.Email, err)
		return metadata.NewListMembershipsInternalServerError()
	}

	var out []*metadata.ListMembershipsOKBodyMembershipsItems0
	for _, om := range oms {
		out = append(out, &metadata.ListMembershipsOKBodyMembershipsItems0{OrganizationToken: om.Token, Description: om.Description, Admin: om.Admin})
	}
	return metadata.NewListMembershipsOK().WithPayload(&metadata.ListMembershipsOKBody{Memberships: out})
}
