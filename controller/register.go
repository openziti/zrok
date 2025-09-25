package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/controller/config"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/account"
)

type registerHandler struct {
	cfg *config.Config
}

func newRegisterHandler(cfg *config.Config) *registerHandler {
	return &registerHandler{
		cfg: cfg,
	}
}
func (h *registerHandler) Handle(params account.RegisterParams) middleware.Responder {
	if params.Body.RegisterToken == "" || params.Body.Password == "" {
		dl.Error("missing token or password")
		return account.NewRegisterNotFound()
	}
	dl.Infof("received register request for registration token '%v'", params.Body.RegisterToken)

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction for registration token '%v': %v", params.Body.RegisterToken, err)
		return account.NewRegisterInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	ar, err := str.FindAccountRequestWithToken(params.Body.RegisterToken, trx)
	if err != nil {
		dl.Errorf("error finding account request with registration token '%v': %v", params.Body.RegisterToken, err)
		return account.NewRegisterNotFound()
	}

	accountToken, err := CreateToken()
	if err != nil {
		dl.Errorf("error creating account token for request '%v' (%v): %v", params.Body.RegisterToken, ar.Email, err)
		return account.NewRegisterInternalServerError()
	}

	if err := validatePassword(h.cfg, params.Body.Password); err != nil {
		dl.Errorf("password not valid for request '%v', (%v): %v", params.Body.RegisterToken, ar.Email, err)
		return account.NewRegisterUnprocessableEntity().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}

	hpwd, err := HashPassword(params.Body.Password)
	if err != nil {
		dl.Errorf("error hashing password for request '%v' (%v): %v", params.Body.RegisterToken, ar.Email, err)
		return account.NewRegisterInternalServerError()
	}
	a := &store.Account{
		Email:    ar.Email,
		Salt:     hpwd.Salt,
		Password: hpwd.Password,
		Token:    accountToken,
	}
	if _, err := str.CreateAccount(a, trx); err != nil {
		dl.Errorf("error creating account for request '%v' (%v): %v", params.Body.RegisterToken, ar.Email, err)
		return account.NewRegisterInternalServerError()
	}

	if err := str.DeleteAccountRequest(ar.Id, trx); err != nil {
		dl.Errorf("error deleteing account request '%v' (%v): %v", params.Body.RegisterToken, ar.Email, err)
		return account.NewRegisterInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		dl.Errorf("error committing '%v' (%v): %v", params.Body.RegisterToken, ar.Email, err)
		return account.NewRegisterInternalServerError()
	}

	dl.Infof("created account '%v'", a.Email)

	return account.NewRegisterOK().WithPayload(&account.RegisterOKBody{AccountToken: a.Token})
}
