package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/controller/store"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/account"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/invite"
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

	if self.cfg.Registration.TokenStrategy == "store" {
		invite, err := str.GetInviteByToken(params.Body.Token, tx)
		if err != nil {
			logrus.Error(err)
			return account.NewInviteBadRequest()
		}
		if err := str.DeleteInvite(invite.Id, tx); err != nil {
			logrus.Error(err)
			return account.NewInviteInternalServerError()
		}
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

type inviteGenerateHandler struct {
}

func newInviteGenerateHandler() *inviteGenerateHandler {
	return &inviteGenerateHandler{}
}

func (handler *inviteGenerateHandler) Handle(params invite.InviteGenerateParams) middleware.Responder {
	if params.Body == nil || len(params.Body.Tokens) == 0 {
		logrus.Error("missing tokens")
		return invite.NewInviteGenerateBadRequest()
	}
	logrus.Infof("received invite generate request with %d tokens", len(params.Body.Tokens))

	invites := make([]*store.Invite, len(params.Body.Tokens))
	for i, token := range params.Body.Tokens {
		invites[i] = &store.Invite{
			Token: token,
		}
	}
	tx, err := str.Begin()
	if err != nil {
		logrus.Error(err)
		return invite.NewInviteGenerateInternalServerError()
	}
	defer func() { _ = tx.Rollback() }()

	if err := str.CreateInvites(invites, tx); err != nil {
		logrus.Error(err)
		return invite.NewInviteGenerateInternalServerError()
	}

	if err := tx.Commit(); err != nil {
		logrus.Errorf("error committing inviteGenerate request: %v", err)
		return account.NewInviteInternalServerError()
	}

	return invite.NewInviteGenerateCreated()
}
