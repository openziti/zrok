package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/share"
	"github.com/sirupsen/logrus"
)

type listShareNamesHandler struct{}

func newListShareNamesHandler() *listShareNamesHandler {
	return &listShareNamesHandler{}
}

func (h *listShareNamesHandler) Handle(params share.ListShareNamesParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return share.NewListShareNamesInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	// find namespace
	ns, err := str.FindNamespaceWithToken(params.NamespaceToken, trx)
	if err != nil {
		logrus.Errorf("error finding namespace with token '%v': %v", params.NamespaceToken, err)
		return share.NewListShareNamesNotFound()
	}

	if !ns.Open {
		// check namespace grant
		granted, err := str.CheckNamespaceGrant(ns.Id, int(principal.ID), trx)
		if err != nil {
			logrus.Errorf("error checking namespace grant for account '%v' and namespace '%v': %v", principal.Email, ns.Token, err)
			return share.NewListShareNamesInternalServerError()
		}
		if !granted {
			logrus.Errorf("account '%v' is not granted access to namespace '%v'", principal.Email, ns.Token)
			return share.NewListShareNamesUnauthorized()
		}
	}

	// find allocated names for namespace
	allocatedNames, err := str.FindNamesForAccountAndNamespace(int(principal.ID), ns.Id, trx)
	if err != nil {
		logrus.Errorf("error finding allocated names for namespace '%v': %v", ns.Token, err)
		return share.NewListShareNamesInternalServerError()
	}

	// build response
	var names []*share.ListShareNamesOKBodyItems0
	for _, an := range allocatedNames {
		nameObj := &share.ListShareNamesOKBodyItems0{
			Name:      an.Name,
			CreatedAt: an.CreatedAt.Unix(),
		}
		names = append(names, nameObj)
	}

	logrus.Debugf("listed %d allocated names for namespace '%v' for account '%v'", len(names), ns.Token, principal.Email)
	return share.NewListShareNamesOK().WithPayload(names)
}
