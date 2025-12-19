package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/automation"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/admin"
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

	for _, env := range envs {
		dl.Infof("disabling environment '%d' (envZId: '%s') for account '%s'", env.Id, env.ZId, params.Body.Email)
		if err := disableEnvironment(env, trx, ziti); err != nil {
			dl.Errorf("error disabling environment '%d' for account '%s': %v", env.Id, params.Body.Email, err)
			return admin.NewDeleteAccountInternalServerError()
		}
		dl.Infof("successfully disabled environment '%d' for account '%s'", env.Id, params.Body.Email)
	}

	if err := str.DeleteAccount(account.Id, trx); err != nil {
		dl.Errorf("error deleting account '%s': %v", params.Body.Email, err)
		return admin.NewDeleteAccountInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		dl.Errorf("error committing transaction: %v", err)
		return admin.NewDeleteAccountInternalServerError()
	}

	dl.Infof("successfully deleted account '%s'", params.Body.Email)
	return admin.NewDeleteAccountOK()
}
