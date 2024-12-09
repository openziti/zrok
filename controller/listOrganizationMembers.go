package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/admin"
	"github.com/sirupsen/logrus"
)

type listOrganizationMembersHandler struct{}

func newListOrganizationMembersHandler() *listOrganizationMembersHandler {
	return &listOrganizationMembersHandler{}
}

func (h *listOrganizationMembersHandler) Handle(params admin.ListOrganizationMembersParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		logrus.Error("invalid admin principal")
		return admin.NewListOrganizationMembersUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return admin.NewListOrganizationMembersInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	org, err := str.FindOrganizationByToken(params.Body.Token, trx)
	if err != nil {
		logrus.Errorf("error finding organization by token: %v", err)
		return admin.NewListOrganizationMembersInternalServerError()
	}
	if org == nil {
		logrus.Errorf("organization '%v' not found", params.Body.Token)
		return admin.NewListOrganizationMembersNotFound()
	}

	emails, err := str.FindAccountsForOrganization(org.Id, trx)
	if err != nil {
		logrus.Errorf("error finding accounts for organization: %v", err)
		return admin.NewListOrganizationMembersInternalServerError()
	}

	var members []*admin.ListOrganizationMembersOKBodyMembersItems0
	for _, email := range emails {
		members = append(members, &admin.ListOrganizationMembersOKBodyMembersItems0{Email: email})
	}
	return admin.NewListOrganizationMembersOK().WithPayload(&admin.ListOrganizationMembersOKBody{Members: members})
}
