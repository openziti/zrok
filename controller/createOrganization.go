package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/admin"
	"github.com/sirupsen/logrus"
)

type createOrganizationHandler struct{}

func newCreateOrganizationHandler() *createOrganizationHandler {
	return &createOrganizationHandler{}
}

func (h *createOrganizationHandler) Handle(params admin.CreateOrganizationParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		logrus.Errorf("invalid admin principal")
		return admin.NewCreateOrganizationUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return admin.NewCreateOrganizationInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	orgToken, err := CreateToken()
	if err != nil {
		logrus.Errorf("error creating organization token: %v", err)
		return admin.NewCreateOrganizationInternalServerError()
	}

	org := &store.Organization{
		Token:       orgToken,
		Description: params.Body.Description,
	}
	if _, err := str.CreateOrganization(org, trx); err != nil {
		logrus.Errorf("error creating organization: %v", err)
		return admin.NewCreateOrganizationInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		logrus.Errorf("error committing organization: %v", err)
		return admin.NewCreateOrganizationInternalServerError()
	}

	logrus.Infof("added organzation '%v' with description '%v'", org.Token, org.Description)

	return admin.NewCreateOrganizationCreated().WithPayload(&admin.CreateOrganizationCreatedBody{OrganizationToken: org.Token})
}
