package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/account"
	"github.com/sirupsen/logrus"
)

type regenerateTokenHandler struct{}

func newRegenerateTokenHandler() *regenerateTokenHandler {
	return &regenerateTokenHandler{}
}

func (handler *regenerateTokenHandler) Handle(params account.RegenerateTokenParams, principal *rest_model_zrok.Principal) middleware.Responder {
	logrus.Infof("received token regeneration request for email '%v'", principal.Email)

	if params.Body.EmailAddress != principal.Email {
		logrus.Errorf("mismatched account '%v' for '%v'", params.Body.EmailAddress, principal.Email)
		return account.NewRegenerateTokenNotFound()
	}

	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction for '%v': %v", params.Body.EmailAddress, err)
		return account.NewRegenerateTokenInternalServerError()
	}
	defer tx.Rollback()

	a, err := str.FindAccountWithEmail(params.Body.EmailAddress, tx)
	if err != nil {
		logrus.Errorf("error finding account for '%v': %v", params.Body.EmailAddress, err)
		return account.NewRegenerateTokenNotFound()
	}
	if a.Deleted {
		logrus.Errorf("account '%v' for '%v' deleted", a.Email, a.Token)
		return account.NewRegenerateTokenNotFound()
	}
	if a.Disabled {
		logrus.Errorf("account '%v' for '%v' disabled", a.Email, a.Token)
		return account.NewResetTokenNotFound()
	}

	// Need to create new token and invalidate all other resources
	token, err := CreateToken()
	if err != nil {
		logrus.Errorf("error creating token for request '%v': %v", params.Body.EmailAddress, err)
		return account.NewRegenerateTokenInternalServerError()
	}

	a.Token = token

	if _, err := str.UpdateAccount(a, tx); err != nil {
		logrus.Errorf("error updating account for request '%v': %v", params.Body.EmailAddress, err)
		return account.NewRegenerateTokenInternalServerError()
	}

	if err := tx.Commit(); err != nil {
		logrus.Errorf("error committing '%v' (%v): %v", params.Body.EmailAddress, a.Email, err)
		return account.NewRegenerateTokenInternalServerError()
	}

	logrus.Infof("regenerated token '%v' for '%v'", a.Token, a.Email)

	return account.NewRegenerateTokenOK().WithPayload(&account.RegenerateTokenOKBody{Token: token})
}
