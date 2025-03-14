package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/admin"
	"github.com/sirupsen/logrus"
)

type removeOrganizationMemberHandler struct{}

func newRemoveOrganizationMemberHandler() *removeOrganizationMemberHandler {
	return &removeOrganizationMemberHandler{}
}

func (h *removeOrganizationMemberHandler) Handle(params admin.RemoveOrganizationMemberParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		logrus.Error("invalid admin principal")
		return admin.NewRemoveOrganizationMemberUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return admin.NewRemoveOrganizationMemberInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	acct, err := str.FindAccountWithEmail(params.Body.Email, trx)
	if err != nil {
		logrus.Errorf("error finding account with email address '%v': %v", params.Body.Email, err)
		return admin.NewAddOrganizationMemberNotFound()
	}

	org, err := str.FindOrganizationByToken(params.Body.OrganizationToken, trx)
	if err != nil {
		logrus.Errorf("error finding organization '%v': %v", params.Body.OrganizationToken, err)
		return admin.NewAddOrganizationMemberNotFound()
	}

	if err := str.RemoveAccountFromOrganization(acct.Id, org.Id, trx); err != nil {
		logrus.Errorf("error removing account '%v' from organization '%v': %v", acct.Email, org.Token, err)
		return admin.NewRemoveOrganizationMemberInternalServerError()
	}

	logrus.Infof("removed '%v' from organization '%v'", acct.Email, org.Token)

	if err := trx.Commit(); err != nil {
		logrus.Errorf("error committing transaction: %v", err)
		return admin.NewRemoveOrganizationMemberInternalServerError()
	}

	return admin.NewRemoveOrganizationMemberOK()
}
