package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/admin"
)

type createNamespaceHandler struct{}

func newCreateNamespaceHandler() *createNamespaceHandler {
	return &createNamespaceHandler{}
}

func (h *createNamespaceHandler) Handle(params admin.CreateNamespaceParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		dl.Errorf("invalid admin principal")
		return admin.NewCreateNamespaceUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return admin.NewCreateNamespaceInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	// check if namespace already exists
	if params.Body.Name != "" {
		if _, err := str.FindNamespaceWithName(params.Body.Name, trx); err == nil {
			dl.Errorf("namespace '%v' already exists", params.Body.Name)
			return admin.NewCreateNamespaceConflict()
		}
	}

	var namespaceToken string
	if params.Body.Token != "" {
		namespaceToken = params.Body.Token
	} else {
		namespaceToken, err = CreateToken()
		if err != nil {
			dl.Errorf("error creating namespace token: %v", err)
			return admin.NewCreateNamespaceInternalServerError()
		}
	}

	ns := &store.Namespace{
		Token:       namespaceToken,
		Name:        params.Body.Name,
		Description: params.Body.Description,
		Open:        params.Body.Open,
	}
	if _, err := str.CreateNamespace(ns, trx); err != nil {
		dl.Errorf("error creating namespace: %v", err)
		return admin.NewCreateNamespaceInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		dl.Errorf("error committing namespace: %v", err)
		return admin.NewCreateNamespaceInternalServerError()
	}

	dl.Infof("added namespace '%v' with name '%v'", ns.Token, ns.Name)

	return admin.NewCreateNamespaceCreated().WithPayload(&admin.CreateNamespaceCreatedBody{NamespaceToken: ns.Token})
}
