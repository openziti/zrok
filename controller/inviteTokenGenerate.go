package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/account"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/admin"
)

type inviteTokenGenerateHandler struct {
}

func newInviteTokenGenerateHandler() *inviteTokenGenerateHandler {
	return &inviteTokenGenerateHandler{}
}

func (handler *inviteTokenGenerateHandler) Handle(params admin.InviteTokenGenerateParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		dl.Errorf("invalid admin principal")
		return admin.NewInviteTokenGenerateUnauthorized()
	}

	if len(params.Body.InviteTokens) == 0 {
		dl.Error("missing tokens")
		return admin.NewInviteTokenGenerateBadRequest()
	}
	dl.Infof("received invite generate request with %d tokens", len(params.Body.InviteTokens))

	invites := make([]*store.InviteToken, len(params.Body.InviteTokens))
	for i, token := range params.Body.InviteTokens {
		invites[i] = &store.InviteToken{
			Token: token,
		}
	}
	trx, err := str.Begin()
	if err != nil {
		dl.Error(err)
		return admin.NewInviteTokenGenerateInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	if err := str.CreateInviteTokens(invites, trx); err != nil {
		dl.Error(err)
		return admin.NewInviteTokenGenerateInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		dl.Errorf("error committing inviteGenerate request: %v", err)
		return account.NewInviteInternalServerError()
	}

	return admin.NewInviteTokenGenerateCreated()
}
