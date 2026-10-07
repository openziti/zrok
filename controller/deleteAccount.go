package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/automation"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/admin"
	"github.com/pkg/errors"
)

type deleteAccountHandler struct{}

func newDeleteAccountHandler() *deleteAccountHandler {
	return &deleteAccountHandler{}
}

func (h *deleteAccountHandler) Handle(params admin.DeleteAccountParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		dl.Error("invalid admin principal")
		return admin.NewDeleteAccountUnauthorized()
	}

	dl.Infof("starting deletion of account with email '%s'", params.Body.Email)

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return admin.NewDeleteAccountInternalServerError()
	}
	defer trx.Rollback()

	account, err := str.FindAccountWithEmail(params.Body.Email, trx)
	if err != nil {
		dl.Errorf("error finding account with email '%s': %v", params.Body.Email, err)
		return admin.NewDeleteAccountNotFound()
	}

	envs, err := str.FindEnvironmentsForAccount(account.Id, trx)
	if err != nil {
		dl.Errorf("error finding environments for account '%s': %v", params.Body.Email, err)
		return admin.NewDeleteAccountInternalServerError()
	}
	dl.Infof("found %d environments to clean up for account '%s'", len(envs), params.Body.Email)

	ziti, err := automation.NewZitiAutomation(cfg.Ziti)
	if err != nil {
		dl.Errorf("error getting automation client: %v", err)
		return admin.NewDeleteAccountInternalServerError()
	}

	var updates []pendingMappingUpdate
	for _, env := range envs {
		dl.Infof("disabling environment '%d' (envZId: '%s') for account '%s'", env.Id, env.ZId, params.Body.Email)
		envUpdates, err := disableEnvironment(env, trx, ziti)
		if err != nil {
			dl.Errorf("error disabling environment '%d' for account '%s': %v", env.Id, params.Body.Email, err)
			return admin.NewDeleteAccountInternalServerError()
		}
		updates = append(updates, envUpdates...)
		dl.Infof("successfully disabled environment '%d' for account '%s'", env.Id, params.Body.Email)
	}

	if err := releaseAccountNames(account.Id, trx); err != nil {
		dl.Errorf("error releasing names for account '%s': %v", params.Body.Email, err)
		return admin.NewDeleteAccountInternalServerError()
	}

	if err := str.DeleteAccount(account.Id, trx); err != nil {
		dl.Errorf("error deleting account '%s': %v", params.Body.Email, err)
		return admin.NewDeleteAccountInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		dl.Errorf("error committing transaction: %v", err)
		return admin.NewDeleteAccountInternalServerError()
	}
	publishMappingUpdates(updates)

	dl.Infof("successfully deleted account '%s'", params.Body.Email)
	return admin.NewDeleteAccountOK()
}

// releaseAccountNames soft-deletes every live name of an account being deleted, reserved names included.
// it runs after the account's environments are torn down, so no live share can still hold one of its
// names; a mapping to a live share is a store error and fails the delete. a mapping to a share deleted
// before teardownShare released names is severed, and is released with its name.
func releaseAccountNames(accountId int, trx *sqlx.Tx) error {
	mappings, err := str.FindShareNameMappingsForAccountNamesWithShare(accountId, trx)
	if err != nil {
		return err
	}
	for _, mapping := range mappings {
		if !mapping.ShareDeleted {
			return errors.Errorf("name mapping '%d' still holds a name for live share '%v'", mapping.Id, mapping.ShareToken)
		}
	}
	for _, mapping := range mappings {
		if err := str.DeleteShareNameMapping(mapping.Id, trx); err != nil {
			return errors.Wrapf(err, "error releasing severed name mapping '%d'", mapping.Id)
		}
	}
	released, err := str.DeleteNamesForAccount(accountId, trx)
	if err != nil {
		return err
	}
	dl.Infof("released '%d' names and '%d' severed name mappings for account '%d'", released, len(mappings), accountId)
	return nil
}
