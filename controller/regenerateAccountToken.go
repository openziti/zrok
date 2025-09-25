package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/account"
)

type regenerateAccountTokenHandler struct{}

func newRegenerateAccountTokenHandler() *regenerateAccountTokenHandler {
	return &regenerateAccountTokenHandler{}
}

func (handler *regenerateAccountTokenHandler) Handle(params account.RegenerateAccountTokenParams, principal *rest_model_zrok.Principal) middleware.Responder {
	dl.Infof("received account token regeneration request for email '%v'", principal.Email)

	if params.Body.EmailAddress != principal.Email {
		dl.Errorf("mismatched account '%v' for '%v'", params.Body.EmailAddress, principal.Email)
		return account.NewRegenerateAccountTokenNotFound()
	}

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction for '%v': %v", params.Body.EmailAddress, err)
		return account.NewRegenerateAccountTokenInternalServerError()
	}
	defer trx.Rollback()

	a, err := str.FindAccountWithEmail(params.Body.EmailAddress, trx)
	if err != nil {
		dl.Errorf("error finding account for '%v': %v", params.Body.EmailAddress, err)
		return account.NewRegenerateAccountTokenNotFound()
	}
	if a.Deleted {
		dl.Errorf("account '%v' for '%v' deleted", a.Email, a.Token)
		return account.NewRegenerateAccountTokenNotFound()
	}

	// Need to create new token and invalidate all other resources
	accountToken, err := CreateToken()
	if err != nil {
		dl.Errorf("error creating account token for request '%v': %v", params.Body.EmailAddress, err)
		return account.NewRegenerateAccountTokenInternalServerError()
	}

	a.Token = accountToken

	if _, err := str.UpdateAccount(a, trx); err != nil {
		dl.Errorf("error updating account for request '%v': %v", params.Body.EmailAddress, err)
		return account.NewRegenerateAccountTokenInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		dl.Errorf("error committing '%v' (%v): %v", params.Body.EmailAddress, a.Email, err)
		return account.NewRegenerateAccountTokenInternalServerError()
	}

	dl.Infof("regenerated account token '%v' for '%v'", a.Token, a.Email)

	return account.NewRegenerateAccountTokenOK().WithPayload(&account.RegenerateAccountTokenOKBody{AccountToken: accountToken})
}
