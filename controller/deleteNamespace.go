package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/admin"
)

type deleteNamespaceHandler struct{}

func newDeleteNamespaceHandler() *deleteNamespaceHandler {
	return &deleteNamespaceHandler{}
}

func (h *deleteNamespaceHandler) Handle(params admin.DeleteNamespaceParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		dl.Errorf("invalid admin principal")
		return admin.NewDeleteNamespaceUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return admin.NewDeleteNamespaceInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	ns, err := str.FindNamespaceWithToken(params.Body.NamespaceToken, trx)
	if err != nil {
		dl.Errorf("error finding namespace by token: %v", err)
		return admin.NewDeleteNamespaceNotFound()
	}

	err = str.DeleteNamespace(ns.Id, trx)
	if err != nil {
		dl.Errorf("error deleting namespace: %v", err)
		return admin.NewDeleteNamespaceInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		dl.Errorf("error committing transaction: %v", err)
		return admin.NewDeleteNamespaceInternalServerError()
	}

	dl.Infof("deleted namespace '%v'", ns.Token)

	return admin.NewDeleteNamespaceOK()
}
