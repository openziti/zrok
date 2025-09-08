package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/share"
	"github.com/sirupsen/logrus"
)

type listNamesForNamespaceHandler struct{}

func newListNamesForNamespaceHandler() *listNamesForNamespaceHandler {
	return &listNamesForNamespaceHandler{}
}

func (h *listNamesForNamespaceHandler) Handle(params share.ListNamesForNamespaceParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return share.NewListNamesForNamespaceInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	// find namespace
	ns, err := str.FindNamespaceWithToken(params.NamespaceToken, trx)
	if err != nil {
		logrus.Errorf("error finding namespace with token '%v': %v", params.NamespaceToken, err)
		return share.NewListNamesForNamespaceNotFound()
	}

	if !ns.Open {
		// check namespace grant
		granted, err := str.CheckNamespaceGrant(ns.Id, int(principal.ID), trx)
		if err != nil {
			logrus.Errorf("error checking namespace grant for account '%v' and namespace '%v': %v", principal.Email, ns.Token, err)
			return share.NewListNamesForNamespaceInternalServerError()
		}
		if !granted {
			logrus.Errorf("account '%v' is not granted access to namespace '%v'", principal.Email, ns.Token)
			return share.NewListNamesForNamespaceUnauthorized()
		}
	}

	// find allocated names for namespace
	names, err := str.FindNamesWithShareTokensForAccountAndNamespace(int(principal.ID), ns.Id, trx)
	if err != nil {
		logrus.Errorf("error finding names for namespace '%v': %v", ns.Token, err)
		return share.NewListNamesForNamespaceInternalServerError()
	}

	// build response
	var out []*rest_model_zrok.Name
	for _, an := range names {
		nameObj := &rest_model_zrok.Name{
			NamespaceToken: ns.Token,
			NamespaceName:  ns.Name,
			Name:           an.Name.Name,
			Reserved:       an.Name.Reserved,
			CreatedAt:      an.Name.CreatedAt.Unix(),
		}
		if an.ShareToken != nil {
			nameObj.ShareToken = *an.ShareToken
		}
		out = append(out, nameObj)
	}

	logrus.Debugf("listed %d names for namespace '%v' for account '%v'", len(out), ns.Token, principal.Email)
	return share.NewListNamesForNamespaceOK().WithPayload(out)
}
