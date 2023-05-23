package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/controller/config"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/account"
	"github.com/sirupsen/logrus"
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
	if params.Body == nil || params.Body.Token == "" || params.Body.Password == "" {
		logrus.Error("missing token or password")
		return account.NewRegisterNotFound()
	}
	logrus.Infof("received register request for token '%v'", params.Body.Token)

	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction for token '%v': %v", params.Body.Token, err)
		return account.NewRegisterInternalServerError()
	}
	defer func() { _ = tx.Rollback() }()

	ar, err := str.FindAccountRequestWithToken(params.Body.Token, tx)
	if err != nil {
		logrus.Errorf("error finding account request with token '%v': %v", params.Body.Token, err)
		return account.NewRegisterNotFound()
	}

	token, err := createToken()
	if err != nil {
		logrus.Errorf("error creating token for request '%v' (%v): %v", params.Body.Token, ar.Email, err)
		return account.NewRegisterInternalServerError()
	}

	if err := validatePassword(h.cfg, params.Body.Password); err != nil {
		logrus.Errorf("password not valid for request '%v', (%v): %v", params.Body.Token, ar.Email, err)
		return account.NewRegisterUnprocessableEntity().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}

	hpwd, err := hashPassword(params.Body.Password)
	if err != nil {
		logrus.Errorf("error hashing password for request '%v' (%v): %v", params.Body.Token, ar.Email, err)
		return account.NewRegisterInternalServerError()
	}
	a := &store.Account{
		Email:    ar.Email,
		Salt:     hpwd.Salt,
		Password: hpwd.Password,
		Token:    token,
	}
	if _, err := str.CreateAccount(a, tx); err != nil {
		logrus.Errorf("error creating account for request '%v' (%v): %v", params.Body.Token, ar.Email, err)
		return account.NewRegisterInternalServerError()
	}

	if err := str.DeleteAccountRequest(ar.Id, tx); err != nil {
		logrus.Errorf("error deleteing account request '%v' (%v): %v", params.Body.Token, ar.Email, err)
		return account.NewRegisterInternalServerError()
	}

	if err := tx.Commit(); err != nil {
		logrus.Errorf("error committing '%v' (%v): %v", params.Body.Token, ar.Email, err)
		return account.NewRegisterInternalServerError()
	}

	logrus.Infof("created account '%v' with token '%v'", a.Email, a.Token)

	return account.NewRegisterOK().WithPayload(&rest_model_zrok.RegisterResponse{Token: a.Token})
}
