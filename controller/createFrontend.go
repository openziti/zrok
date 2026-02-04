package controller

import (
	"errors"

	"github.com/go-openapi/runtime/middleware"
	"github.com/lib/pq"
	"github.com/mattn/go-sqlite3"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/automation"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/admin"
)

type createFrontendHandler struct{}

func newCreateFrontendHandler() *createFrontendHandler {
	return &createFrontendHandler{}
}

func (h *createFrontendHandler) Handle(params admin.CreateFrontendParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		dl.Errorf("invalid admin principal")
		return admin.NewCreateFrontendUnauthorized()
	}

	ziti, err := automation.NewZitiAutomation(cfg.Ziti)
	if err != nil {
		dl.Errorf("error getting automation client: %v", err)
		return admin.NewCreateFrontendInternalServerError()
	}

	zId := params.Body.ZID
	identity, err := ziti.Identities.GetByID(zId)
	if err != nil {
		dl.Errorf("error getting identity details for '%v': %v", zId, err)
		if ziti.IsNotFound(err) {
			return admin.NewCreateFrontendNotFound()
		}
		return admin.NewCreateFrontendInternalServerError()
	}
	dl.Infof("found frontend identity '%v'", *identity.Name)

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return admin.NewCreateFrontendInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	feToken, err := CreateToken()
	if err != nil {
		dl.Errorf("error creating frontend token: %v", err)
		return admin.NewCreateFrontendInternalServerError()
	}

	fe := &store.Frontend{
		Token:          feToken,
		ZId:            params.Body.ZID,
		PublicName:     &params.Body.PublicName,
		UrlTemplate:    &params.Body.URLTemplate,
		Reserved:       true,
		PermissionMode: store.PermissionMode(params.Body.PermissionMode),
		Dynamic:        params.Body.Dynamic,
	}
	if _, err := str.CreateGlobalFrontend(fe, trx); err != nil {
		perr := &pq.Error{}
		sqliteErr := &sqlite3.Error{}
		switch {
		case errors.As(err, &perr):
			if perr.Code == pq.ErrorCode("23505") {
				dl.Errorf("error creating frontend record: %v", err)
				return admin.NewCreateFrontendBadRequest()
			}
		case errors.As(err, sqliteErr):
			if errors.Is(sqliteErr.Code, sqlite3.ErrConstraint) {
				dl.Errorf("error creating frontend record: %v", err)
				return admin.NewCreateFrontendBadRequest()
			}
		}

		dl.Errorf("error creating frontend record: %v", err)
		return admin.NewCreateFrontendInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		dl.Errorf("error committing frontend record: %v", err)
		return admin.NewCreateFrontendInternalServerError()
	}

	dl.Infof("created global frontend '%v' with public name '%v'", fe.Token, *fe.PublicName)

	return admin.NewCreateFrontendCreated().WithPayload(&admin.CreateFrontendCreatedBody{FrontendToken: feToken})
}
