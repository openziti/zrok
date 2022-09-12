package controller

import (
	"crypto/sha512"
	"encoding/hex"
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/controller/store"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/identity"
	"github.com/sirupsen/logrus"
)

type createAccountHandler struct {
	cfg *Config
}

func newCreateAccountHandler(cfg *Config) *createAccountHandler {
	return &createAccountHandler{cfg: cfg}
}

func (self *createAccountHandler) Handle(params identity.CreateAccountParams) middleware.Responder {
	logrus.Infof("received account request for email '%v'", params.Body.Email)
	if self.cfg.Registration.ImmediateCreate {
		return self.handleDirectCreate(params)
	} else {
		return self.handleVerifiedCreate(params)
	}
}

func (self *createAccountHandler) handleDirectCreate(params identity.CreateAccountParams) middleware.Responder {
	if params.Body == nil || params.Body.Email == "" || params.Body.Password == "" {
		logrus.Errorf("missing email or password")
		return identity.NewCreateAccountBadRequest().WithPayload("missing email or password")
	}

	token, err := generateApiToken()
	if err != nil {
		logrus.Errorf("error generating api token: %v", err)
		return identity.NewCreateAccountInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	a := &store.Account{
		Email:    params.Body.Email,
		Password: hashPassword(params.Body.Password),
		Token:    token,
	}
	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return identity.NewCreateAccountInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	id, err := str.CreateAccount(a, tx)
	if err != nil {
		logrus.Errorf("error creating account: %v", err)
		_ = tx.Rollback()
		return identity.NewCreateAccountBadRequest().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	if err := tx.Commit(); err != nil {
		logrus.Errorf("error comitting: %v", err)
	}

	logrus.Infof("account created with id = '%v'", id)
	return identity.NewCreateAccountCreated().WithPayload(&rest_model_zrok.AccountResponse{Token: token})
}

func (self *createAccountHandler) handleVerifiedCreate(params identity.CreateAccountParams) middleware.Responder {
	if params.Body == nil || params.Body.Email == "" {
		logrus.Errorf("missing email")
		return identity.NewCreateAccountBadRequest().WithPayload("missing email")
	}
	token, err := generateApiToken()
	if err != nil {
		logrus.Errorf("error generating api token: %v", err)
		return identity.NewCreateAccountInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	if err := sendVerificationEmail(params.Body.Email, token, self.cfg); err != nil {
		logrus.Error(err)
		return identity.NewCreateAccountInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	return identity.NewCreateAccountCreated()
}

func hashPassword(raw string) string {
	hash := sha512.New()
	hash.Write([]byte(raw))
	return hex.EncodeToString(hash.Sum(nil))
}
