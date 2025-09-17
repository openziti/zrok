package controller

import (
	"errors"

	"github.com/go-openapi/runtime/middleware"
	"github.com/lib/pq"
	"github.com/mattn/go-sqlite3"
	"github.com/openziti/zrok/controller/automation"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/admin"
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

	ziti, err := automation.NewZitiAutomation(cfg)
	if err != nil {
		logrus.Errorf("error getting automation client: %v", err)
		return admin.NewCreateFrontendInternalServerError()
	}

	zId := params.Body.ZID
	identity, err := ziti.Identities.GetByID(zId)
	if err != nil {
		logrus.Errorf("error getting identity details for '%v': %v", zId, err)
		if ziti.IsNotFound(err) {
			return admin.NewCreateFrontendNotFound()
		}
		return admin.NewCreateFrontendInternalServerError()
	}
	logrus.Infof("found frontend identity '%v'", *identity.Name)

	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return admin.NewCreateFrontendInternalServerError()
	}
	defer func() { _ = tx.Rollback() }()

	feToken, err := CreateToken()
	if err != nil {
		logrus.Errorf("error creating frontend token: %v", err)
		return admin.NewCreateFrontendInternalServerError()
	}

	fe := &store.Frontend{
		Token:          feToken,
		ZId:            params.Body.ZID,
		PublicName:     &params.Body.PublicName,
		UrlTemplate:    &params.Body.URLTemplate,
		Reserved:       true,
		PermissionMode: store.PermissionMode(params.Body.PermissionMode),
	}
	if _, err := str.CreateGlobalFrontend(fe, tx); err != nil {
		perr := &pq.Error{}
		sqliteErr := &sqlite3.Error{}
		switch {
		case errors.As(err, &perr):
			if perr.Code == pq.ErrorCode("23505") {
				return admin.NewCreateFrontendBadRequest()
			}
		case errors.As(err, sqliteErr):
			if errors.Is(sqliteErr.Code, sqlite3.ErrConstraint) {
				return admin.NewCreateFrontendBadRequest()
			}
		}

		logrus.Errorf("error creating frontend record: %v", err)
		return admin.NewCreateFrontendInternalServerError()
	}

	if err := tx.Commit(); err != nil {
		logrus.Errorf("error committing frontend record: %v", err)
		return admin.NewCreateFrontendInternalServerError()
	}

	logrus.Infof("created global frontend '%v' with public name '%v'", fe.Token, *fe.PublicName)

	return admin.NewCreateFrontendCreated().WithPayload(&admin.CreateFrontendCreatedBody{FrontendToken: feToken})
}
