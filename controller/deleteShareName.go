package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/share"
	"github.com/sirupsen/logrus"
)

type deleteShareNameHandler struct{}

func newDeleteShareNameHandler() *deleteShareNameHandler {
	return &deleteShareNameHandler{}
}

func (h *deleteShareNameHandler) Handle(params share.DeleteShareNameParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return share.NewDeleteShareNameInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	// find namespace
	ns, err := str.FindNamespaceWithToken(params.Body.NamespaceToken, trx)
	if err != nil {
		logrus.Errorf("error finding namespace with token '%v': %v", params.Body.NamespaceToken, err)
		return share.NewDeleteShareNameNotFound()
	}

	if !ns.Open {
		// check namespace grant
		granted, err := str.CheckNamespaceGrant(ns.Id, int(principal.ID), trx)
		if err != nil {
			logrus.Errorf("error checking namespace grant for account '%v' and namespace '%v': %v", principal.Email, ns.Token, err)
			return share.NewDeleteShareNameInternalServerError()
		}
		if !granted {
			logrus.Errorf("account '%v' is not granted access to namespace '%v'", principal.Email, ns.Token)
			return share.NewDeleteShareNameUnauthorized()
		}
	}

	// find allocated name
	an, err := str.FindAllocatedNameByNamespaceAndName(ns.Id, params.Body.Name, trx)
	if err != nil {
		logrus.Errorf("error finding allocated name '%v' in namespace '%v': %v", params.Body.Name, ns.Token, err)
		return share.NewDeleteShareNameNotFound()
	}

	// verify ownership
	if an.AccountId != int(principal.ID) {
		logrus.Errorf("account '%v' does not own name '%v' in namespace '%v'", principal.Email, params.Body.Name, ns.Token)
		return share.NewDeleteShareNameUnauthorized()
	}

	// delete allocated name
	if err := str.DeleteAllocatedName(an.Id, trx); err != nil {
		logrus.Errorf("error deleting allocated name '%v' in namespace '%v' for account '%v': %v", params.Body.Name, ns.Token, principal.Email, err)
		return share.NewDeleteShareNameInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		logrus.Errorf("error committing transaction: %v", err)
		return share.NewDeleteShareNameInternalServerError()
	}

	logrus.Infof("deleted allocated name '%v' in namespace '%v' for account '%v'", params.Body.Name, ns.Token, principal.Email)
	return share.NewDeleteShareNameOK()
}
