package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/admin"
)

type removeNamespaceGrantHandler struct{}

func newRemoveNamespaceGrantHandler() *removeNamespaceGrantHandler {
	return &removeNamespaceGrantHandler{}
}

func (h *removeNamespaceGrantHandler) Handle(params admin.RemoveNamespaceGrantParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		dl.Error("invalid admin principal")
		return admin.NewRemoveNamespaceGrantUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return admin.NewRemoveNamespaceGrantInternalServerError()
	}
	defer trx.Rollback()

	ns, err := str.FindNamespaceWithToken(params.Body.NamespaceToken, trx)
	if err != nil {
		dl.Errorf("error finding namespace with token '%v': %v", params.Body.NamespaceToken, err)
		return admin.NewRemoveNamespaceGrantNotFound()
	}

	acct, err := str.FindAccountWithEmail(params.Body.Email, trx)
	if err != nil {
		dl.Errorf("error finding account with email '%v': %v", params.Body.Email, err)
		return admin.NewRemoveNamespaceGrantNotFound()
	}

	if granted, err := str.CheckNamespaceGrant(ns.Id, acct.Id, trx); err != nil {
		dl.Errorf("error checking namespace grant for account '%v' and namespace '%v': %v", acct.Email, ns.Token, err)
		return admin.NewRemoveNamespaceGrantInternalServerError()

	} else if granted {
		grants, err := str.FindNamespaceGrantsForNamespace(ns.Id, trx)
		if err != nil {
			dl.Errorf("error finding grants for namespace '%v': %v", ns.Token, err)
			return admin.NewRemoveNamespaceGrantInternalServerError()
		}

		var grantId int
		for _, grant := range grants {
			if grant.AccountId == acct.Id {
				grantId = grant.Id
				break
			}
		}

		if grantId == 0 {
			dl.Errorf("grant not found for account '%v' and namespace '%v'", acct.Email, ns.Token)
			return admin.NewRemoveNamespaceGrantNotFound()
		}

		if err := str.DeleteNamespaceGrant(grantId, trx); err != nil {
			dl.Errorf("error deleting namespace ('%v') grant for '%v': %v", ns.Token, acct.Email, err)
			return admin.NewRemoveNamespaceGrantInternalServerError()
		}
		dl.Infof("removed '%v' access to namespace '%v'", acct.Email, ns.Token)

		if err := trx.Commit(); err != nil {
			dl.Errorf("error committing transaction: %v", err)
			return admin.NewRemoveNamespaceGrantInternalServerError()
		}

	} else {
		dl.Infof("account '%v' not granted access to namespace '%v'", acct.Email, ns.Token)
	}

	return admin.NewRemoveNamespaceGrantOK()
}
