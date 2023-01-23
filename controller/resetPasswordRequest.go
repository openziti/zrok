package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/rest_server_zrok/operations/account"
	"github.com/openziti/zrok/util"
	"github.com/sirupsen/logrus"
)

type resetPasswordRequestHandler struct{}

func newResetPasswordRequestHandler() *resetPasswordRequestHandler {
	return &resetPasswordRequestHandler{}
}

func (handler *resetPasswordRequestHandler) Handle(params account.ResetPasswordRequestParams) middleware.Responder {
	if params.Body.EmailAddress == "" {
		logrus.Errorf("missing email")
		return account.NewResetPasswordRequestBadRequest()
	}
	if !util.IsValidEmail(params.Body.EmailAddress) {
		logrus.Errorf("'%v' is not a valid email address", params.Body.EmailAddress)
		return account.NewResetPasswordRequestBadRequest()
	}
	logrus.Infof("received reset password request for email '%v'", params.Body.EmailAddress)
	var token string

	tx, err := str.Begin()
	if err != nil {
		logrus.Error(err)
		return account.NewResetPasswordRequestInternalServerError()
	}
	defer func() { _ = tx.Rollback() }()

	token, err = createToken()
	if err != nil {
		logrus.Error(err)
		return account.NewResetPasswordRequestInternalServerError()
	}

	acct, err := str.FindAccountWithEmail(params.Body.EmailAddress, tx)
	if err != nil {
		logrus.Infof("no account found for '%v': %v", params.Body.EmailAddress, err)
		return account.NewResetPasswordRequestInternalServerError()
	}

	prr := &store.PasswordResetRequest{
		Token:     token,
		AccountId: acct.Id,
	}

	if _, err := str.CreatePasswordResetRequest(prr, tx); err != nil {
		logrus.Errorf("error creating reset password request for '%v': %v", params.Body.EmailAddress, err)
		return account.NewResetPasswordRequestInternalServerError()
	}

	if err := tx.Commit(); err != nil {
		logrus.Errorf("error committing reset password request for '%v': %v", params.Body.EmailAddress, err)
		return account.NewResetPasswordRequestInternalServerError()
	}

	if cfg.Email != nil && cfg.Registration != nil && cfg.ResetPassword != nil {
		if err := sendResetPasswordEmail(acct.Email, token); err != nil {
			logrus.Errorf("error sending reset password email for '%v': %v", acct.Email, err)
			return account.NewResetPasswordRequestInternalServerError()
		}
	} else {
		logrus.Errorf("'email', 'registration', and 'reset_password' configuration missing; skipping reset password email")
	}

	logrus.Infof("reset password request for '%v' has token '%v'", params.Body.EmailAddress, prr.Token)

	return account.NewResetPasswordRequestCreated()
}
