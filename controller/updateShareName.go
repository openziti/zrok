package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/share"
	"github.com/sirupsen/logrus"
)

type updateShareNameHandler struct{}

func newUpdateShareNameHandler() *updateShareNameHandler {
	return &updateShareNameHandler{}
}

func (h *updateShareNameHandler) Handle(params share.UpdateShareNameParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return share.NewUpdateShareNameInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	// find namespace
	ns, err := str.FindNamespaceWithToken(params.Body.NamespaceToken, trx)
	if err != nil {
		logrus.Errorf("error finding namespace with token '%v': %v", params.Body.NamespaceToken, err)
		return share.NewUpdateShareNameNotFound()
	}

	// check namespace grant
	if !ns.Open {
		granted, err := str.CheckNamespaceGrant(ns.Id, int(principal.ID), trx)
		if err != nil {
			logrus.Errorf("error checking namespace grant for account '%v' and namespace '%v': %v", principal.Email, ns.Token, err)
			return share.NewUpdateShareNameInternalServerError()
		}
		if !granted {
			logrus.Errorf("account '%v' is not granted access to namespace '%v'", principal.Email, ns.Token)
			return share.NewUpdateShareNameUnauthorized()
		}
	}

	// find existing name
	name, err := str.FindNameByNamespaceAndName(ns.Id, params.Body.Name, trx)
	if err != nil {
		logrus.Errorf("error finding name '%v' in namespace '%v': %v", params.Body.Name, ns.Token, err)
		return share.NewUpdateShareNameNotFound()
	}

	// verify ownership
	if name.AccountId != int(principal.ID) {
		logrus.Errorf("account '%v' does not own name '%v' in namespace '%v'", principal.Email, params.Body.Name, ns.Token)
		return share.NewUpdateShareNameUnauthorized()
	}

	// check if update is actually changing something
	if name.Reserved == params.Body.Reserved {
		logrus.Debugf("no change needed for name '%v' in namespace '%v' - already has reserved=%v", params.Body.Name, ns.Token, params.Body.Reserved)
		return share.NewUpdateShareNameOK()
	}

	// update the reservation state
	name.Reserved = params.Body.Reserved
	if err := str.UpdateName(name, trx); err != nil {
		logrus.Errorf("error updating name '%v' in namespace '%v' for account '%v': %v", params.Body.Name, ns.Token, principal.Email, err)
		return share.NewUpdateShareNameInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		logrus.Errorf("error committing transaction: %v", err)
		return share.NewUpdateShareNameInternalServerError()
	}

	logrus.Infof("updated name '%v' in namespace '%v' for account '%v' - reserved set to %v", params.Body.Name, ns.Token, principal.Email, params.Body.Reserved)
	return share.NewUpdateShareNameOK()
}
