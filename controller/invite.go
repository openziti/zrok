package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/controller/store"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/account"
	"github.com/openziti-test-kitchen/zrok/util"
	"github.com/sirupsen/logrus"
)

type inviteHandler struct {
	cfg *Config
}

func newInviteHandler(cfg *Config) *inviteHandler {
	return &inviteHandler{
		cfg: cfg,
	}
}

func (self *inviteHandler) Handle(params account.InviteParams) middleware.Responder {
	if params.Body == nil || params.Body.Email == "" {
		logrus.Errorf("missing email")
		return account.NewInviteBadRequest()
	}
	if !util.IsValidEmail(params.Body.Email) {
		logrus.Errorf("'%v' is not a valid email address", params.Body.Email)
		return account.NewInviteBadRequest()
	}
	logrus.Infof("received account request for email '%v'", params.Body.Email)
	var token string

	tx, err := str.Begin()
	if err != nil {
		logrus.Error(err)
		return account.NewInviteInternalServerError()
	}
	defer func() { _ = tx.Rollback() }()

	if self.cfg.Registration != nil && self.cfg.Registration.TokenStrategy == "store" {
		inviteToken, err := str.GetInviteTokenByToken(params.Body.Token, tx)
		if err != nil {
			logrus.Errorf("cannot get invite token '%v' for '%v': %v", params.Body.Token, params.Body.Email, err)
			return account.NewInviteBadRequest()
		}
		if err := str.DeleteInviteToken(inviteToken.Id, tx); err != nil {
			logrus.Error(err)
			return account.NewInviteInternalServerError()
		}
		logrus.Infof("using invite token '%v' to process invite request for '%v'", inviteToken.Token, params.Body.Email)
	}

	token, err = createToken()
	if err != nil {
		logrus.Error(err)
		return account.NewInviteInternalServerError()
	}
	ar := &store.AccountRequest{
		Token:         token,
		Email:         params.Body.Email,
		SourceAddress: params.HTTPRequest.RemoteAddr,
	}

	if _, err := str.FindAccountWithEmail(params.Body.Email, tx); err == nil {
		logrus.Errorf("found account for '%v', cannot process account request", params.Body.Email)
		return account.NewInviteBadRequest()
	} else {
		logrus.Infof("no account found for '%v': %v", params.Body.Email, err)
	}

	if oldAr, err := str.FindAccountRequestWithEmail(params.Body.Email, tx); err == nil {
		logrus.Warnf("found previous account request for '%v', removing", params.Body.Email)
		if err := str.DeleteAccountRequest(oldAr.Id, tx); err != nil {
			logrus.Errorf("error deleteing previous account request for '%v': %v", params.Body.Email, err)
			return account.NewInviteInternalServerError()
		}
	} else {
		logrus.Warnf("error finding previous account request for '%v': %v", params.Body.Email, err)
	}

	if _, err := str.CreateAccountRequest(ar, tx); err != nil {
		logrus.Errorf("error creating account request for '%v': %v", params.Body.Email, err)
		return account.NewInviteInternalServerError()
	}
	if err := tx.Commit(); err != nil {
		logrus.Errorf("error committing account request for '%v': %v", params.Body.Email, err)
		return account.NewInviteInternalServerError()
	}

	if cfg.Email != nil && cfg.Registration != nil {
		if err := sendVerificationEmail(params.Body.Email, token); err != nil {
			logrus.Errorf("error sending verification email for '%v': %v", params.Body.Email, err)
			return account.NewInviteInternalServerError()
		}
	} else {
		logrus.Errorf("'email' and 'registration' configuration missing; skipping registration email")
	}

	logrus.Infof("account request for '%v' has registration token '%v'", params.Body.Email, ar.Token)

	return account.NewInviteCreated()
}
