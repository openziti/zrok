package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/account"
	"github.com/sirupsen/logrus"
)

type regenerateAccountTokenHandler struct{}

func newRegenerateAccountTokenHandler() *regenerateAccountTokenHandler {
	return &regenerateAccountTokenHandler{}
}

func (handler *regenerateAccountTokenHandler) Handle(params account.RegenerateAccountTokenParams, principal *rest_model_zrok.Principal) middleware.Responder {
	logrus.Infof("received account token regeneration request for email '%v'", principal.Email)

	if params.Body.EmailAddress != principal.Email {
		logrus.Errorf("mismatched account '%v' for '%v'", params.Body.EmailAddress, principal.Email)
		return account.NewRegenerateAccountTokenNotFound()
	}

	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction for '%v': %v", params.Body.EmailAddress, err)
		return account.NewRegenerateAccountTokenInternalServerError()
	}
	defer tx.Rollback()

	a, err := str.FindAccountWithEmail(params.Body.EmailAddress, tx)
	if err != nil {
		logrus.Errorf("error finding account for '%v': %v", params.Body.EmailAddress, err)
		return account.NewRegenerateAccountTokenNotFound()
	}
	if a.Deleted {
		logrus.Errorf("account '%v' for '%v' deleted", a.Email, a.Token)
		return account.NewRegenerateAccountTokenNotFound()
	}

	// Need to create new token and invalidate all other resources
	accountToken, err := CreateToken()
	if err != nil {
		logrus.Errorf("error creating account token for request '%v': %v", params.Body.EmailAddress, err)
		return account.NewRegenerateAccountTokenInternalServerError()
	}

	a.Token = accountToken

	if _, err := str.UpdateAccount(a, tx); err != nil {
		logrus.Errorf("error updating account for request '%v': %v", params.Body.EmailAddress, err)
		return account.NewRegenerateAccountTokenInternalServerError()
	}

	if err := tx.Commit(); err != nil {
		logrus.Errorf("error committing '%v' (%v): %v", params.Body.EmailAddress, a.Email, err)
		return account.NewRegenerateAccountTokenInternalServerError()
	}

	logrus.Infof("regenerated account token '%v' for '%v'", a.Token, a.Email)

	return account.NewRegenerateAccountTokenOK().WithPayload(&account.RegenerateAccountTokenOKBody{AccountToken: accountToken})
}
