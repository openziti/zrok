package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/admin"
)

type addNamespaceGrantHandler struct{}

func newAddNamespaceGrantHandler() *addNamespaceGrantHandler {
	return &addNamespaceGrantHandler{}
}

func (h *addNamespaceGrantHandler) Handle(params admin.AddNamespaceGrantParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		dl.Error("invalid admin principal")
		return admin.NewAddNamespaceGrantUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return admin.NewAddNamespaceGrantInternalServerError()
	}
	defer trx.Rollback()

	ns, err := str.FindNamespaceWithToken(params.Body.NamespaceToken, trx)
	if err != nil {
		dl.Errorf("error finding namespace with token '%v': %v", params.Body.NamespaceToken, err)
		return admin.NewAddNamespaceGrantNotFound()
	}

	acct, err := str.FindAccountWithEmail(params.Body.Email, trx)
	if err != nil {
		dl.Errorf("error finding account with email '%v': %v", params.Body.Email, err)
		return admin.NewAddNamespaceGrantNotFound()
	}

	if granted, err := str.CheckNamespaceGrant(ns.Id, acct.Id, trx); err != nil {
		dl.Errorf("error checking namespace grant for account '%v' and namespace '%v': %v", acct.Email, ns.Token, err)
		return admin.NewAddNamespaceGrantInternalServerError()

	} else if !granted {
		ng := &store.NamespaceGrant{
			NamespaceId: ns.Id,
			AccountId:   acct.Id,
		}
		if _, err := str.CreateNamespaceGrant(ng, trx); err != nil {
			dl.Errorf("error creating namespace ('%v') grant for '%v': %v", ns.Token, acct.Email, err)
			return admin.NewAddNamespaceGrantInternalServerError()
		}
		dl.Infof("granted '%v' access to namespace '%v'", acct.Email, ns.Token)

		if err := trx.Commit(); err != nil {
			dl.Errorf("error committing transaction: %v", err)
			return admin.NewAddNamespaceGrantInternalServerError()
		}

	} else {
		dl.Infof("account '%v' already granted access to namespace '%v'", acct.Email, ns.Token)
	}

	return admin.NewAddNamespaceGrantOK()
}
