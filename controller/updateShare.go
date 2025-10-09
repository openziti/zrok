package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/share"
)

type updateShareHandler struct{}

func newUpdateShareHandler() *updateShareHandler {
	return &updateShareHandler{}
}

func (h *updateShareHandler) Handle(params share.UpdateShareParams, principal *rest_model_zrok.Principal) middleware.Responder {
	shrToken := params.Body.ShareToken

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return share.NewUpdateShareInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	sshr, err := str.FindShareWithToken(shrToken, trx)
	if err != nil {
		dl.Errorf("share '%v' not found: %v", shrToken, err)
		return share.NewUpdateShareNotFound()
	}

	senvs, err := str.FindEnvironmentsForAccount(int(principal.ID), trx)
	if err != nil {
		dl.Errorf("error finding environments for account '%v': %v", principal.Email, err)
		return share.NewUpdateShareInternalServerError()
	}

	envFound := false
	for _, senv := range senvs {
		if senv.Id == sshr.EnvironmentId {
			envFound = true
			break
		}
	}
	if !envFound {
		dl.Errorf("environment not found for share '%v'", shrToken)
		return share.NewUpdateShareNotFound()
	}

	doCommit := false
	for _, addr := range params.Body.AddAccessGrants {
		acct, err := str.FindAccountWithEmail(addr, trx)
		if err != nil {
			dl.Errorf("error looking up account by email '%v' for user '%v': %v", addr, principal.Email, err)
			return share.NewUpdateShareBadRequest()
		}
		if _, err := str.CreateAccessGrant(sshr.Id, acct.Id, trx); err != nil {
			dl.Errorf("error adding access grant '%v' for share '%v': %v", acct.Email, shrToken, err)
			return share.NewUpdateShareInternalServerError()
		}
		dl.Infof("added access grant '%v' to share '%v'", acct.Email, shrToken)
		doCommit = true
	}

	for _, addr := range params.Body.RemoveAccessGrants {
		acct, err := str.FindAccountWithEmail(addr, trx)
		if err != nil {
			dl.Errorf("error looking up account by email '%v' for user '%v': %v", addr, principal.Email, err)
			return share.NewUpdateShareBadRequest()
		}
		if err := str.DeleteAccessGrantsForShareAndAccount(sshr.Id, acct.Id, trx); err != nil {
			dl.Errorf("error removing access grant '%v' for share '%v': %v", acct.Email, shrToken, err)
			return share.NewUpdateShareInternalServerError()
		}
		dl.Infof("removed access grant '%v' from share '%v'", acct.Email, shrToken)
		doCommit = true
	}

	if doCommit {
		if err := trx.Commit(); err != nil {
			dl.Errorf("error committing transaction for share '%v' update: %v", shrToken, err)
			return share.NewUpdateShareInternalServerError()
		}
	}

	return share.NewUpdateShareOK()
}
