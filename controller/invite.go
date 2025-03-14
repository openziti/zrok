package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/controller/config"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/rest_server_zrok/operations/account"
	"github.com/openziti/zrok/util"
	"github.com/sirupsen/logrus"
)

type inviteHandler struct {
	cfg *config.Config
}

func newInviteHandler(cfg *config.Config) *inviteHandler {
	return &inviteHandler{
		cfg: cfg,
	}
}

func (h *inviteHandler) Handle(params account.InviteParams) middleware.Responder {
	if h.cfg.Invites == nil || !h.cfg.Invites.InvitesOpen {
		logrus.Warnf("not accepting invites; attempt from '%v'", params.Body.Email)
		return account.NewInviteBadRequest()
	}
	if params.Body.Email == "" {
		logrus.Errorf("missing email")
		return account.NewInviteBadRequest()
	}
	if !util.IsValidEmail(params.Body.Email) {
		logrus.Errorf("'%v' is not a valid email address", params.Body.Email)
		return account.NewInviteBadRequest()
	}
	logrus.Infof("received account request for email '%v'", params.Body.Email)
	var regToken string

	tx, err := str.Begin()
	if err != nil {
		logrus.Error(err)
		return account.NewInviteInternalServerError()
	}
	defer func() { _ = tx.Rollback() }()

	if h.cfg.Invites != nil && h.cfg.Invites.TokenStrategy == "store" {
		inviteToken, err := str.FindInviteTokenByToken(params.Body.InviteToken, tx)
		if err != nil {
			logrus.Errorf("cannot get invite token '%v' for '%v': %v", params.Body.InviteToken, params.Body.Email, err)
			return account.NewInviteBadRequest().WithPayload("missing invite token")
		}
		if err := str.DeleteInviteToken(inviteToken.Id, tx); err != nil {
			logrus.Error(err)
			return account.NewInviteInternalServerError()
		}
		logrus.Infof("using invite token '%v' to process invite request for '%v'", inviteToken.Token, params.Body.Email)
	}

	regToken, err = CreateToken()
	if err != nil {
		logrus.Error(err)
		return account.NewInviteInternalServerError()
	}
	ar := &store.AccountRequest{
		Token:         regToken,
		Email:         params.Body.Email,
		SourceAddress: params.HTTPRequest.RemoteAddr,
	}

	// deleted accounts still exist as far as invites are concerned (ignore deleted flag)
	if _, err := str.FindAccountWithEmailAndDeleted(params.Body.Email, tx); err == nil {
		logrus.Errorf("found account for '%v', cannot process account request", params.Body.Email)
		return account.NewInviteBadRequest().WithPayload("duplicate email found")
	} else {
		logrus.Infof("no account found for '%v': %v", params.Body.Email, err)
	}

	if oldAr, err := str.FindAccountRequestWithEmail(params.Body.Email, tx); err == nil {
		logrus.Warnf("found previous account request for '%v', removing", params.Body.Email)
		if err := str.DeleteAccountRequest(oldAr.Id, tx); err != nil {
			logrus.Errorf("error deleting previous account request for '%v': %v", params.Body.Email, err)
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
		if err := sendVerificationEmail(params.Body.Email, regToken); err != nil {
			logrus.Errorf("error sending verification email for '%v': %v", params.Body.Email, err)
			return account.NewInviteInternalServerError()
		}
	} else {
		logrus.Errorf("'email' and 'registration' configuration missing; skipping registration email")
	}

	logrus.Infof("account request for '%v' has registration token '%v'", params.Body.Email, ar.Token)

	return account.NewInviteCreated()
}
