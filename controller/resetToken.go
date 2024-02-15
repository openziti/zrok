package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/account"
	"github.com/sirupsen/logrus"
)

type resetTokenHandler struct{}

func newResetTokenHandler() *resetTokenHandler {
	return &resetTokenHandler{}
}

func (handler *resetTokenHandler) Handle(params account.ResetTokenParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if params.Body.EmailAddress == "" {
		logrus.Error("missing email")
		return account.NewResetTokenNotFound()
	}
	logrus.Infof("received token reset request for email '%v'", params.Body.EmailAddress)

	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction for '%v': %v", params.Body.EmailAddress, err)
		return account.NewResetTokenInternalServerError()
	}
	defer tx.Rollback()

	a, err := str.FindAccountWithEmail(params.Body.EmailAddress, tx)
	if err != nil {
		logrus.Errorf("error finding account for '%v': %v", params.Body.EmailAddress, err)
		return account.NewResetTokenNotFound()
	}
	if a.Deleted {
		logrus.Errorf("account '%v' for '%v' deleted", a.Email, a.Token)
		return account.NewResetTokenNotFound()
	}

	// Need to create new token and invalidate all other resources
	token, err := CreateToken()
	if err != nil {
		logrus.Errorf("error creating token for request '%v': %v", params.Body.EmailAddress, err)
		return account.NewResetTokenInternalServerError()
	}

	a.Token = token

	if _, err := str.UpdateAccount(a, tx); err != nil {
		logrus.Errorf("error updating account for request '%v': %v", params.Body.EmailAddress, err)
		return account.NewResetTokenInternalServerError()
	}

	if err := tx.Commit(); err != nil {
		logrus.Errorf("error committing '%v' (%v): %v", params.Body.EmailAddress, a.Email, err)
		return account.NewResetTokenInternalServerError()
	}

	logrus.Infof("reset token for '%v'", a.Email)

	return account.NewResetTokenOK().WithPayload(&account.ResetTokenOKBody{Token: token})
}
