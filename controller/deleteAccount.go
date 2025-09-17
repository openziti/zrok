package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/controller/automation"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/admin"
	"github.com/sirupsen/logrus"
)

type deleteAccountHandler struct{}

func newDeleteAccountHandler() *deleteAccountHandler {
	return &deleteAccountHandler{}
}

func (h *deleteAccountHandler) Handle(params admin.DeleteAccountParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		logrus.Error("invalid admin principal")
		return admin.NewDeleteAccountUnauthorized()
	}

	logrus.Infof("starting deletion of account with email '%s'", params.Body.Email)

	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return admin.NewDeleteAccountInternalServerError()
	}
	defer trx.Rollback()

	account, err := str.FindAccountWithEmail(params.Body.Email, trx)
	if err != nil {
		logrus.Errorf("error finding account with email '%s': %v", params.Body.Email, err)
		return admin.NewDeleteAccountNotFound()
	}

	envs, err := str.FindEnvironmentsForAccount(account.Id, trx)
	if err != nil {
		logrus.Errorf("error finding environments for account '%s': %v", params.Body.Email, err)
		return admin.NewDeleteAccountInternalServerError()
	}
	logrus.Infof("found %d environments to clean up for account '%s'", len(envs), params.Body.Email)

	ziti, err := automation.NewZitiAutomation(cfg.Ziti)
	if err != nil {
		logrus.Errorf("error getting automation client: %v", err)
		return admin.NewDeleteAccountInternalServerError()
	}

	for _, env := range envs {
		logrus.Infof("disabling environment '%d' (envZId: '%s') for account '%s'", env.Id, env.ZId, params.Body.Email)
		if err := disableEnvironment(env, trx, ziti); err != nil {
			logrus.Errorf("error disabling environment '%d' for account '%s': %v", env.Id, params.Body.Email, err)
			return admin.NewDeleteAccountInternalServerError()
		}
		logrus.Infof("successfully disabled environment '%d' for account '%s'", env.Id, params.Body.Email)
	}

	if err := str.DeleteAccount(account.Id, trx); err != nil {
		logrus.Errorf("error deleting account '%s': %v", params.Body.Email, err)
		return admin.NewDeleteAccountInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		logrus.Errorf("error committing transaction: %v", err)
		return admin.NewDeleteAccountInternalServerError()
	}

	logrus.Infof("successfully deleted account '%s'", params.Body.Email)
	return admin.NewDeleteAccountOK()
}
