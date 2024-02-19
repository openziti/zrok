package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/controller/config"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/account"
	"github.com/sirupsen/logrus"
)

type resetPasswordHandler struct {
	cfg *config.Config
}

func newResetPasswordHandler(cfg *config.Config) *resetPasswordHandler {
	return &resetPasswordHandler{
		cfg: cfg,
	}
}

func (handler *resetPasswordHandler) Handle(params account.ResetPasswordParams) middleware.Responder {
	if params.Body == nil || params.Body.Token == "" || params.Body.Password == "" {
		logrus.Error("missing token or password")
		return account.NewResetPasswordNotFound()
	}
	logrus.Infof("received password reset request for token '%v'", params.Body.Token)

	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction for '%v': %v", params.Body.Token, err)
		return account.NewResetPasswordInternalServerError()
	}
	defer func() { _ = tx.Rollback() }()

	prr, err := str.FindPasswordResetRequestWithToken(params.Body.Token, tx)
	if err != nil {
		logrus.Errorf("error finding reset request for '%v': %v", params.Body.Token, err)
		return account.NewResetPasswordNotFound()
	}

	a, err := str.GetAccount(prr.AccountId, tx)
	if err != nil {
		logrus.Errorf("error finding account for '%v': %v", params.Body.Token, err)
		return account.NewResetPasswordNotFound()
	}
	if a.Deleted {
		logrus.Errorf("account '%v' for '%v' deleted", a.Email, a.Token)
		return account.NewResetPasswordNotFound()
	}
	if a.Disabled {
		logrus.Errorf("account '%v' for '%v' disabled", a.Email, a.Token)
		return account.NewResetPasswordNotFound()
	}

	if err := validatePassword(handler.cfg, params.Body.Password); err != nil {
		logrus.Errorf("password not valid for request '%v', (%v): %v", params.Body.Token, a.Email, err)
		return account.NewResetPasswordUnprocessableEntity().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}

	hpwd, err := HashPassword(params.Body.Password)
	if err != nil {
		logrus.Errorf("error hashing password for '%v' (%v): %v", params.Body.Token, a.Email, err)
		return account.NewResetPasswordRequestInternalServerError()
	}
	a.Salt = hpwd.Salt
	a.Password = hpwd.Password

	if _, err := str.UpdateAccount(a, tx); err != nil {
		logrus.Errorf("error updating for '%v' (%v): %v", params.Body.Token, a.Email, err)
		return account.NewResetPasswordInternalServerError()
	}

	if err := str.DeletePasswordResetRequest(prr.Id, tx); err != nil {
		logrus.Errorf("error deleting reset request '%v' (%v): %v", params.Body.Token, a.Email, err)
		return account.NewResetPasswordInternalServerError()
	}

	if err := tx.Commit(); err != nil {
		logrus.Errorf("error committing '%v' (%v): %v", params.Body.Token, a.Email, err)
		return account.NewResetPasswordInternalServerError()
	}

	logrus.Infof("reset password for '%v'", a.Email)

	return account.NewResetPasswordOK()
}
