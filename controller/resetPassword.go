package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/config"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/account"
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
	if params.Body.ResetToken == "" || params.Body.Password == "" {
		dl.Error("missing token or password")
		return account.NewResetPasswordNotFound()
	}
	dl.Infof("received password reset request for reset token '%v'", params.Body.ResetToken)

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction for '%v': %v", params.Body.ResetToken, err)
		return account.NewResetPasswordInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	prr, err := str.FindPasswordResetRequestWithToken(params.Body.ResetToken, trx)
	if err != nil {
		dl.Errorf("error finding reset request for reset token '%v': %v", params.Body.ResetToken, err)
		return account.NewResetPasswordNotFound()
	}

	a, err := str.GetAccount(prr.AccountId, trx)
	if err != nil {
		dl.Errorf("error finding account for reset token '%v': %v", params.Body.ResetToken, err)
		return account.NewResetPasswordNotFound()
	}
	if a.Deleted {
		dl.Errorf("account '%v' for '%v' deleted", a.Email, a.Token)
		return account.NewResetPasswordNotFound()
	}

	if err := validatePassword(handler.cfg, params.Body.Password); err != nil {
		dl.Errorf("password not valid for reset token '%v', (%v): %v", params.Body.ResetToken, a.Email, err)
		return account.NewResetPasswordUnprocessableEntity().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}

	hpwd, err := HashPassword(params.Body.Password)
	if err != nil {
		dl.Errorf("error hashing password for '%v' (%v): %v", params.Body.ResetToken, a.Email, err)
		return account.NewResetPasswordRequestInternalServerError()
	}
	a.Salt = hpwd.Salt
	a.Password = hpwd.Password

	if _, err := str.UpdateAccount(a, trx); err != nil {
		dl.Errorf("error updating for reset token '%v' (%v): %v", params.Body.ResetToken, a.Email, err)
		return account.NewResetPasswordInternalServerError()
	}

	if err := str.DeletePasswordResetRequest(prr.Id, trx); err != nil {
		dl.Errorf("error deleting reset request for reset token '%v' (%v): %v", params.Body.ResetToken, a.Email, err)
		return account.NewResetPasswordInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		dl.Errorf("error committing for reset token '%v' (%v): %v", params.Body.ResetToken, a.Email, err)
		return account.NewResetPasswordInternalServerError()
	}

	dl.Infof("reset password for '%v'", a.Email)

	return account.NewResetPasswordOK()
}
