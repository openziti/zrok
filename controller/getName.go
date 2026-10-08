package controller

import (
	"database/sql"
	"errors"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/admin"
)

type getNameHandler struct{}

func newGetNameHandler() *getNameHandler {
	return &getNameHandler{}
}

func (h *getNameHandler) Handle(params admin.GetNameParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		dl.Error("invalid admin principal")
		return admin.NewGetNameUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return admin.NewGetNameInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	ns, err := str.FindNamespaceWithToken(params.NamespaceToken, trx)
	if errors.Is(err, sql.ErrNoRows) {
		return admin.NewGetNameNotFound()
	}
	if err != nil {
		dl.Errorf("error finding namespace with token '%s': %v", params.NamespaceToken, err)
		return admin.NewGetNameInternalServerError()
	}

	an, err := str.FindNameWithShareTokenByNamespaceAndName(ns.Id, params.Name, trx)
	if errors.Is(err, sql.ErrNoRows) {
		return admin.NewGetNameNotFound()
	}
	if err != nil {
		dl.Errorf("error finding name '%s' in namespace '%s': %v", params.Name, ns.Token, err)
		return admin.NewGetNameInternalServerError()
	}

	account, err := str.GetAccount(an.AccountId, trx)
	if err != nil {
		dl.Errorf("error finding account for name '%s' in namespace '%s': %v", params.Name, ns.Token, err)
		return admin.NewGetNameInternalServerError()
	}

	out := &admin.GetNameOKBody{
		NamespaceToken: swag.String(ns.Token),
		NamespaceName:  swag.String(ns.Name),
		Name:           swag.String(an.Name.Name),
		AccountEmail:   swag.String(account.Email),
		ShareToken:     swag.String(""),
		Reserved:       swag.Bool(an.Name.Reserved),
		CreatedAt:      swag.Int64(an.Name.CreatedAt.Unix()),
	}
	if an.ShareToken != nil {
		out.ShareToken = an.ShareToken
	}
	return admin.NewGetNameOK().WithPayload(out)
}
