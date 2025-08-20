package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/admin"
	"github.com/sirupsen/logrus"
)

type deleteNamespaceHandler struct{}

func newDeleteNamespaceHandler() *deleteNamespaceHandler {
	return &deleteNamespaceHandler{}
}

func (h *deleteNamespaceHandler) Handle(params admin.DeleteNamespaceParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		logrus.Errorf("invalid admin principal")
		return admin.NewDeleteNamespaceUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return admin.NewDeleteNamespaceInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	ns, err := str.FindNamespaceByToken(params.Body.NamespaceToken, trx)
	if err != nil {
		logrus.Errorf("error finding namespace by token: %v", err)
		return admin.NewDeleteNamespaceNotFound()
	}

	err = str.DeleteNamespace(ns.Id, trx)
	if err != nil {
		logrus.Errorf("error deleting namespace: %v", err)
		return admin.NewDeleteNamespaceInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		logrus.Errorf("error committing transaction: %v", err)
		return admin.NewDeleteNamespaceInternalServerError()
	}

	logrus.Infof("deleted namespace '%v'", ns.Token)

	return admin.NewDeleteNamespaceOK()
}