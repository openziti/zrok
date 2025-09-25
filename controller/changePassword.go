package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/controller/config"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/account"
)

type changePasswordHandler struct {
	cfg *config.Config
}

func newChangePasswordHandler(cfg *config.Config) *changePasswordHandler {
	return &changePasswordHandler{
		cfg: cfg,
	}
}

func (handler *changePasswordHandler) Handle(params account.ChangePasswordParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if params.Body.Email == "" || params.Body.OldPassword == "" || params.Body.NewPassword == "" {
		dl.Error("missing email, old, or new password")
		return account.NewChangePasswordUnauthorized()
	}
	dl.Infof("received change password request for email '%v'", params.Body.Email)

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return account.NewChangePasswordUnauthorized()
	}
	defer func() { _ = trx.Rollback() }()

	a, err := str.FindAccountWithEmail(params.Body.Email, trx)
	if err != nil {
		dl.Errorf("error finding account '%v': %v", params.Body.Email, err)
		return account.NewChangePasswordUnauthorized()
	}
	ohpwd, err := rehashPassword(params.Body.OldPassword, a.Salt)
	if err != nil {
		dl.Errorf("error hashing password for '%v': %v", params.Body.Email, err)
		return account.NewChangePasswordUnauthorized()
	}
	if a.Password != ohpwd.Password {
		dl.Errorf("password mismatch for account '%v'", params.Body.Email)
		return account.NewChangePasswordUnauthorized()
	}

	if err := validatePassword(handler.cfg, params.Body.NewPassword); err != nil {
		dl.Errorf("password not valid for request '%v': %v", a.Email, err)
		return account.NewChangePasswordUnprocessableEntity().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}

	nhpwd, err := HashPassword(params.Body.NewPassword)
	if err != nil {
		dl.Errorf("error hashing password for '%v': %v", a.Email, err)
		return account.NewChangePasswordInternalServerError()
	}
	a.Salt = nhpwd.Salt
	a.Password = nhpwd.Password

	if _, err := str.UpdateAccount(a, trx); err != nil {
		dl.Errorf("error updating for '%v': %v", a.Email, err)
		return account.NewChangePasswordInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		dl.Errorf("error committing '%v': %v", a.Email, err)
		return account.NewChangePasswordInternalServerError()
	}

	dl.Infof("change password for '%v'", a.Email)
	return account.NewChangePasswordOK()
}
