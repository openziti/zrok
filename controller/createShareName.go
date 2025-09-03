package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/share"
	"github.com/sirupsen/logrus"
)

type createShareNameHandler struct{}

func newCreateShareNameHandler() *createShareNameHandler {
	return &createShareNameHandler{}
}

func (h *createShareNameHandler) Handle(params share.CreateShareNameParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return share.NewCreateShareNameInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	// find namespace
	ns, err := str.FindNamespaceWithToken(params.Body.NamespaceToken, trx)
	if err != nil {
		logrus.Errorf("error finding namespace with token '%v': %v", params.Body.NamespaceToken, err)
		return share.NewCreateShareNameNotFound()
	}

	// check namespace grant
	if !ns.Open {
		granted, err := str.CheckNamespaceGrant(ns.Id, int(principal.ID), trx)
		if err != nil {
			logrus.Errorf("error checking namespace grant for account '%v' and namespace '%v': %v", principal.Email, ns.Token, err)
			return share.NewCreateShareNameInternalServerError()
		}
		if !granted {
			logrus.Errorf("account '%v' is not granted access to namespace '%v'", principal.Email, ns.Token)
			return share.NewCreateShareNameUnauthorized()
		}
	}

	// check name availability
	available, err := str.CheckNameAvailability(ns.Id, params.Body.Name, trx)
	if err != nil {
		logrus.Errorf("error checking name availability for '%v' in namespace '%v': %v", params.Body.Name, ns.Token, err)
		return share.NewCreateShareNameInternalServerError()
	}
	if !available {
		logrus.Errorf("name '%v' already exists in namespace '%v'", params.Body.Name, ns.Token)
		return share.NewCreateShareNameConflict()
	}

	// create allocated name
	an := &store.Name{
		NamespaceId: ns.Id,
		Name:        params.Body.Name,
		AccountId:   int(principal.ID),
	}
	_, err = str.CreateName(an, trx)
	if err != nil {
		logrus.Errorf("error creating allocated name '%v' in namespace '%v' for account '%v': %v", params.Body.Name, ns.Token, principal.Email, err)
		return share.NewCreateShareNameInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		logrus.Errorf("error committing transaction: %v", err)
		return share.NewCreateShareNameInternalServerError()
	}

	logrus.Infof("created allocated name '%v' in namespace '%v' for account '%v'", params.Body.Name, ns.Token, principal.Email)
	return share.NewCreateShareNameCreated()
}
