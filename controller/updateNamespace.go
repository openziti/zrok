package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/admin"
	"github.com/sirupsen/logrus"
)

type updateNamespaceHandler struct{}

func newUpdateNamespaceHandler() *updateNamespaceHandler {
	return &updateNamespaceHandler{}
}

func (h *updateNamespaceHandler) Handle(params admin.UpdateNamespaceParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		logrus.Errorf("invalid admin principal")
		return admin.NewUpdateNamespaceUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return admin.NewUpdateNamespaceInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	ns, err := str.FindNamespaceWithToken(params.Body.NamespaceToken, trx)
	if err != nil {
		logrus.Errorf("error finding namespace by token: %v", err)
		return admin.NewUpdateNamespaceNotFound()
	}

	// check if name change conflicts with existing namespace
	if params.Body.Name != "" && params.Body.Name != ns.Name {
		if _, err := str.FindNamespaceWithName(params.Body.Name, trx); err == nil {
			logrus.Errorf("namespace name '%v' already exists", params.Body.Name)
			return admin.NewUpdateNamespaceInternalServerError()
		}
		ns.Name = params.Body.Name
	}

	if params.Body.Description != "" {
		ns.Description = params.Body.Description
	}

	if params.Body.OpenSet {
		ns.Open = params.Body.Open
	}

	if err := str.UpdateNamespace(ns, trx); err != nil {
		logrus.Errorf("error updating namespace: %v", err)
		return admin.NewUpdateNamespaceInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		logrus.Errorf("error committing transaction: %v", err)
		return admin.NewUpdateNamespaceInternalServerError()
	}

	logrus.Infof("updated namespace '%v' with name '%v'", ns.Token, ns.Name)

	return admin.NewUpdateNamespaceOK()
}
