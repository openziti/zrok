package controller

import (
	"fmt"

	"github.com/go-openapi/runtime/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/share"
	"github.com/openziti/zrok/util"
	"github.com/pkg/errors"
)

type createShareNameHandler struct{}

func newCreateShareNameHandler() *createShareNameHandler {
	return &createShareNameHandler{}
}

func (h *createShareNameHandler) Handle(params share.CreateShareNameParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return share.NewCreateShareNameInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	// find namespace
	ns, err := str.FindNamespaceWithToken(params.Body.NamespaceToken, trx)
	if err != nil {
		dl.Errorf("error finding namespace with token '%v': %v", params.Body.NamespaceToken, err)
		return share.NewCreateShareNameNotFound()
	}

	// check namespace grant
	if !ns.Open {
		granted, err := str.CheckNamespaceGrant(ns.Id, int(principal.ID), trx)
		if err != nil {
			dl.Errorf("error checking namespace grant for account '%v' and namespace '%v': %v", principal.Email, ns.Token, err)
			return share.NewCreateShareNameInternalServerError()
		}
		if !granted {
			dl.Errorf("account '%v' is not granted access to namespace '%v'", principal.Email, ns.Token)
			return share.NewCreateShareNameUnauthorized()
		}
	}

	// check limits
	if err := h.checkLimits(principal, trx); err != nil {
		dl.Errorf("limits error: %v", err)
		return share.NewCreateShareNameConflict().WithPayload("names limit reached; cannot reserve additional names")
	}

	// check name availability
	available, err := str.CheckNameAvailability(ns.Id, params.Body.Name, trx)
	if err != nil {
		dl.Errorf("error checking name availability for '%v' in namespace '%v': %v", params.Body.Name, ns.Token, err)
		return share.NewCreateShareNameInternalServerError()
	}
	if !available {
		dl.Errorf("name '%v' already exists in namespace '%v'", params.Body.Name, ns.Token)
		return share.NewCreateShareNameConflict()
	}

	// screen for profanity and DNS-appropriateness
	if !util.IsValidUniqueName(params.Body.Name) {
		dl.Errorf("'%v' is not a valid unique name for '%v'", params.Body.Name, principal.Email)
		return share.NewCreateShareNameConflict().WithPayload(rest_model_zrok.ErrorMessage(fmt.Sprintf("'%v' is not a valid share name; failed profanity or DNS check", params.Body.Name)))
	}

	// create allocated name
	an := &store.Name{
		NamespaceId: ns.Id,
		Name:        params.Body.Name,
		Reserved:    true,
		AccountId:   int(principal.ID),
	}
	_, err = str.CreateName(an, trx)
	if err != nil {
		dl.Errorf("error creating allocated name '%v' in namespace '%v' for account '%v': %v", params.Body.Name, ns.Token, principal.Email, err)
		return share.NewCreateShareNameInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		dl.Errorf("error committing transaction: %v", err)
		return share.NewCreateShareNameInternalServerError()
	}

	dl.Infof("created allocated name '%v' in namespace '%v' for account '%v'", params.Body.Name, ns.Token, principal.Email)
	return share.NewCreateShareNameCreated()
}

func (h *createShareNameHandler) checkLimits(principal *rest_model_zrok.Principal, trx *sqlx.Tx) error {
	if !principal.Limitless {
		if limitsAgent != nil {
			ok, err := limitsAgent.CanReserveName(int(principal.ID), trx)
			if err != nil {
				return errors.Wrapf(err, "error checking name limits for '%v'", principal.Email)
			}
			if !ok {
				return errors.Errorf("name limit check failed for '%v'", principal.Email)
			}
		}
	}
	return nil
}
