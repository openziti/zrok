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

func createAccountHandler(params identity.CreateAccountParams) middleware.Responder {
	logrus.Infof("received account request for username '%v'", params.Body.Username)
	if params.Body == nil || params.Body.Username == "" || params.Body.Password == "" {
		logrus.Errorf("missing username or password")
		return identity.NewCreateAccountBadRequest().WithPayload("missing username or password")
	}

	token, err := generateApiToken()
	if err != nil {
		logrus.Errorf("error generating api token: %v", err)
		return identity.NewCreateAccountInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	a := &store.Account{
		Username: params.Body.Username,
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

func hashPassword(raw string) string {
	hash := sha512.New()
	hash.Write([]byte(raw))
	return hex.EncodeToString(hash.Sum(nil))
}
