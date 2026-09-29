package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/automation"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/share"
	"github.com/pkg/errors"
)

type unshareHandler struct{}

func newUnshareHandler() *unshareHandler {
	return &unshareHandler{}
}

func (h *unshareHandler) Handle(params share.UnshareParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction for '%v': %v", principal.Email, err)
		return share.NewUnshareInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	shrToken := params.Body.ShareToken
	envZId := params.Body.EnvZID

	// validate environment
	env, err := h.validateEnvironment(envZId, principal, trx)
	if err != nil {
		dl.Errorf("environment validation failed for '%v': %v", principal.Email, err)
		return share.NewUnshareNotFound()
	}

	// find and validate share
	shr, err := h.findAndValidateShare(shrToken, env, trx)
	if err != nil {
		dl.Errorf("share validation failed for '%v': %v", principal.Email, err)
		return share.NewUnshareNotFound()
	}

	ziti, err := automation.NewZitiAutomation(cfg.Ziti)
	if err != nil {
		dl.Errorf("error getting automation client for '%v': %v", principal.Email, err)
		return share.NewUnshareInternalServerError()
	}

	updates, err := teardownShare(shr, trx, ziti)
	if err != nil {
		dl.Errorf("error tearing down share '%v' for '%v': %v", shrToken, principal.Email, err)
		return share.NewUnshareInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		dl.Errorf("error committing transaction for '%v': %v", shrToken, err)
		return share.NewUnshareInternalServerError()
	}
	publishMappingUpdates(updates)

	dl.Infof("successfully unshared '%v' for '%v'", shrToken, principal.Email)
	return share.NewUnshareOK()
}

func (h *unshareHandler) validateEnvironment(envZId string, principal *rest_model_zrok.Principal, trx *sqlx.Tx) (*store.Environment, error) {
	env, err := str.FindEnvironmentForAccount(envZId, int(principal.ID), trx)
	if err != nil {
		return nil, errors.Wrapf(err, "error finding environment '%v' for account '%v'", envZId, principal.Email)
	}
	return env, nil
}

func (h *unshareHandler) findAndValidateShare(shrToken string, env *store.Environment, trx *sqlx.Tx) (*store.Share, error) {
	shares, err := str.FindSharesForEnvironment(env.Id, trx)
	if err != nil {
		return nil, errors.Wrapf(err, "error finding shares for environment '%v'", env.ZId)
	}

	for _, share := range shares {
		if share.Token == shrToken {
			return share, nil
		}
	}

	return nil, errors.Errorf("share '%v' not found in environment '%v'", shrToken, env.ZId)
}
