package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/admin"
	"github.com/sirupsen/logrus"
)

type addNamespaceGrantHandler struct{}

func newAddNamespaceGrantHandler() *addNamespaceGrantHandler {
	return &addNamespaceGrantHandler{}
}

func (h *addNamespaceGrantHandler) Handle(params admin.AddNamespaceGrantParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		logrus.Error("invalid admin principal")
		return admin.NewAddNamespaceGrantUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return admin.NewAddNamespaceGrantInternalServerError()
	}
	defer trx.Rollback()

	ns, err := str.FindNamespaceWithToken(params.Body.NamespaceToken, trx)
	if err != nil {
		logrus.Errorf("error finding namespace with token '%v': %v", params.Body.NamespaceToken, err)
		return admin.NewAddNamespaceGrantNotFound()
	}

	acct, err := str.FindAccountWithEmail(params.Body.Email, trx)
	if err != nil {
		logrus.Errorf("error finding account with email '%v': %v", params.Body.Email, err)
		return admin.NewAddNamespaceGrantNotFound()
	}

	if granted, err := str.CheckNamespaceGrant(ns.Id, acct.Id, trx); err != nil {
		logrus.Errorf("error checking namespace grant for account '%v' and namespace '%v': %v", acct.Email, ns.Token, err)
		return admin.NewAddNamespaceGrantInternalServerError()

	} else if !granted {
		ng := &store.NamespaceGrant{
			NamespaceId: ns.Id,
			AccountId:   acct.Id,
		}
		if _, err := str.CreateNamespaceGrant(ng, trx); err != nil {
			logrus.Errorf("error creating namespace ('%v') grant for '%v': %v", ns.Token, acct.Email, err)
			return admin.NewAddNamespaceGrantInternalServerError()
		}
		logrus.Infof("granted '%v' access to namespace '%v'", acct.Email, ns.Token)

		if err := trx.Commit(); err != nil {
			logrus.Errorf("error committing transaction: %v", err)
			return admin.NewAddNamespaceGrantInternalServerError()
		}

	} else {
		logrus.Infof("account '%v' already granted access to namespace '%v'", acct.Email, ns.Token)
	}

	return admin.NewAddNamespaceGrantOK()
}
