package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/controller/store"
	"github.com/openziti-test-kitchen/zrok/controller/zrokEdgeSdk"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/admin"
	"github.com/sirupsen/logrus"
)

type createFrontendHandler struct{}

func newCreateFrontendHandler() *createFrontendHandler {
	return &createFrontendHandler{}
}

func (h *createFrontendHandler) Handle(params admin.CreateFrontendParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		logrus.Errorf("invalid admin principal")
		return admin.NewCreateFrontendUnauthorized()
	}

	client, err := edgeClient()
	if err != nil {
		logrus.Errorf("error getting edge client: %v", err)
		return admin.NewCreateFrontendInternalServerError()
	}

	zId := params.Body.ZID
	detail, err := zrokEdgeSdk.GetIdentity(zId, client)
	if err != nil {
		logrus.Errorf("error getting identity details for '%v': %v", zId, err)
		return admin.NewCreateFrontendInternalServerError()
	}
	if len(detail.Payload.Data) != 1 {
		logrus.Errorf("expected a single identity to be returned for '%v'", zId)
		return admin.NewCreateFrontendNotFound()
	}
	logrus.Infof("found frontend identity '%v'", *detail.Payload.Data[0].Name)

	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return admin.NewCreateFrontendInternalServerError()
	}
	defer func() { _ = tx.Rollback() }()

	feToken, err := createToken()
	if err != nil {
		logrus.Errorf("error creating frontend token: %v", err)
		return admin.NewCreateFrontendInternalServerError()
	}

	fe := &store.Frontend{
		Token:       feToken,
		ZId:         params.Body.ZID,
		PublicName:  &params.Body.PublicName,
		UrlTemplate: &params.Body.URLTemplate,
		Reserved:    true,
	}
	if _, err := str.CreateGlobalFrontend(fe, tx); err != nil {
		logrus.Errorf("error creating frontend record: %v", err)
		return admin.NewCreateFrontendInternalServerError()
	}

	if err := tx.Commit(); err != nil {
		logrus.Errorf("error committing frontend record: %v", err)
		return admin.NewCreateFrontendInternalServerError()
	}

	logrus.Infof("created global frontend '%v' with public name '%v'", fe.Token, *fe.PublicName)

	return admin.NewCreateFrontendCreated().WithPayload(&rest_model_zrok.CreateFrontendResponse{Token: feToken})
}
