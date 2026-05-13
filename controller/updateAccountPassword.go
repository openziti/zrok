package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/config"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/admin"
)

type updateAccountPasswordHandler struct {
	cfg *config.Config
}

func newUpdateAccountPasswordHandler(cfg *config.Config) *updateAccountPasswordHandler {
	return &updateAccountPasswordHandler{
		cfg: cfg,
	}
}

func (handler *updateAccountPasswordHandler) Handle(params admin.UpdateAccountPasswordParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		dl.Errorf("invalid admin principal")
		return admin.NewUpdateAccountPasswordUnauthorized()
	}

	if params.Body.Email == "" || params.Body.Password == "" {
		dl.Error("missing email or password")
		return admin.NewUpdateAccountPasswordNotFound()
	}

	if err := validatePassword(handler.cfg, params.Body.Password); err != nil {
		dl.Errorf("password not valid for request '%v': %v", params.Body.Email, err)
		return admin.NewUpdateAccountPasswordUnprocessableEntity().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return admin.NewUpdateAccountPasswordInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	a, err := str.FindAccountWithEmail(params.Body.Email, trx)
	if err != nil {
		dl.Errorf("error finding account '%v': %v", params.Body.Email, err)
		return admin.NewUpdateAccountPasswordNotFound()
	}

	hpwd, err := HashPassword(params.Body.Password)
	if err != nil {
		dl.Errorf("error hashing password for '%v': %v", params.Body.Email, err)
		return admin.NewUpdateAccountPasswordInternalServerError()
	}
	a.Salt = hpwd.Salt
	a.Password = hpwd.Password

	if _, err := str.UpdateAccount(a, trx); err != nil {
		dl.Errorf("error updating account '%v': %v", params.Body.Email, err)
		return admin.NewUpdateAccountPasswordInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		dl.Errorf("error committing '%v': %v", params.Body.Email, err)
		return admin.NewUpdateAccountPasswordInternalServerError()
	}

	dl.Infof("updated password for '%v'", params.Body.Email)
	return admin.NewUpdateAccountPasswordOK()
}
