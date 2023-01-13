package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/account"
	"github.com/openziti/zrok/rest_server_zrok/operations/admin"
	"github.com/sirupsen/logrus"
)

type inviteTokenGenerateHandler struct {
}

func newInviteTokenGenerateHandler() *inviteTokenGenerateHandler {
	return &inviteTokenGenerateHandler{}
}

func (handler *inviteTokenGenerateHandler) Handle(params admin.InviteTokenGenerateParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		logrus.Errorf("invalid admin principal")
		return admin.NewInviteTokenGenerateUnauthorized()
	}

	if params.Body == nil || len(params.Body.Tokens) == 0 {
		logrus.Error("missing tokens")
		return admin.NewInviteTokenGenerateBadRequest()
	}
	logrus.Infof("received invite generate request with %d tokens", len(params.Body.Tokens))

	invites := make([]*store.InviteToken, len(params.Body.Tokens))
	for i, token := range params.Body.Tokens {
		invites[i] = &store.InviteToken{
			Token: token,
		}
	}
	tx, err := str.Begin()
	if err != nil {
		logrus.Error(err)
		return admin.NewInviteTokenGenerateInternalServerError()
	}
	defer func() { _ = tx.Rollback() }()

	if err := str.CreateInviteTokens(invites, tx); err != nil {
		logrus.Error(err)
		return admin.NewInviteTokenGenerateInternalServerError()
	}

	if err := tx.Commit(); err != nil {
		logrus.Errorf("error committing inviteGenerate request: %v", err)
		return account.NewInviteInternalServerError()
	}

	return admin.NewInviteTokenGenerateCreated()
}
