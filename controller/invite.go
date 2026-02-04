package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/config"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/account"
	"github.com/openziti/zrok/v2/util"
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
		dl.Warnf("not accepting invites; attempt from '%v'", params.Body.Email)
		return account.NewInviteBadRequest()
	}
	if params.Body.Email == "" {
		dl.Errorf("missing email")
		return account.NewInviteBadRequest()
	}
	if !util.IsValidEmail(params.Body.Email) {
		dl.Errorf("'%v' is not a valid email address", params.Body.Email)
		return account.NewInviteBadRequest()
	}
	dl.Infof("received account request for email '%v'", params.Body.Email)
	var regToken string

	trx, err := str.Begin()
	if err != nil {
		dl.Error(err)
		return account.NewInviteInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	if h.cfg.Invites != nil && h.cfg.Invites.TokenStrategy == "store" {
		inviteToken, err := str.FindInviteTokenByToken(params.Body.InviteToken, trx)
		if err != nil {
			dl.Errorf("cannot get invite token '%v' for '%v': %v", params.Body.InviteToken, params.Body.Email, err)
			return account.NewInviteBadRequest().WithPayload("missing invite token")
		}
		if err := str.DeleteInviteToken(inviteToken.Id, trx); err != nil {
			dl.Error(err)
			return account.NewInviteInternalServerError()
		}
		dl.Infof("using invite token '%v' to process invite request for '%v'", inviteToken.Token, params.Body.Email)
	}

	regToken, err = CreateToken()
	if err != nil {
		dl.Error(err)
		return account.NewInviteInternalServerError()
	}
	ar := &store.AccountRequest{
		Token:         regToken,
		Email:         params.Body.Email,
		SourceAddress: params.HTTPRequest.RemoteAddr,
	}

	// deleted accounts still exist as far as invites are concerned (ignore deleted flag)
	if _, err := str.FindAccountWithEmailAndDeleted(params.Body.Email, trx); err == nil {
		dl.Errorf("found account for '%v', cannot process account request", params.Body.Email)
		return account.NewInviteBadRequest().WithPayload("duplicate email found")
	} else {
		dl.Infof("no account found for '%v': %v", params.Body.Email, err)
	}

	if oldAr, err := str.FindAccountRequestWithEmail(params.Body.Email, trx); err == nil {
		dl.Warnf("found previous account request for '%v', removing", params.Body.Email)
		if err := str.DeleteAccountRequest(oldAr.Id, trx); err != nil {
			dl.Errorf("error deleting previous account request for '%v': %v", params.Body.Email, err)
			return account.NewInviteInternalServerError()
		}
	} else {
		dl.Warnf("error finding previous account request for '%v': %v", params.Body.Email, err)
	}

	if _, err := str.CreateAccountRequest(ar, trx); err != nil {
		dl.Errorf("error creating account request for '%v': %v", params.Body.Email, err)
		return account.NewInviteInternalServerError()
	}
	if err := trx.Commit(); err != nil {
		dl.Errorf("error committing account request for '%v': %v", params.Body.Email, err)
		return account.NewInviteInternalServerError()
	}

	if cfg.Email != nil && cfg.Registration != nil {
		if err := sendVerificationEmail(params.Body.Email, regToken); err != nil {
			dl.Errorf("error sending verification email for '%v': %v", params.Body.Email, err)
			return account.NewInviteInternalServerError()
		}
	} else {
		dl.Errorf("'email' and 'registration' configuration missing; skipping registration email")
	}

	dl.Infof("account request for '%v' has registration token '%v'", params.Body.Email, ar.Token)

	return account.NewInviteCreated()
}
