package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/share"
)

type listAllNamesHandler struct{}

func newListAllNamesHandler() *listAllNamesHandler {
	return &listAllNamesHandler{}
}

func (h *listAllNamesHandler) Handle(params share.ListAllNamesParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return share.NewListAllNamesInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	// find all namespaces the user has access to
	namespaces, err := str.FindNamespacesForAccount(int(principal.ID), trx)
	if err != nil {
		dl.Errorf("error finding namespaces for account '%v': %v", principal.Email, err)
		return share.NewListAllNamesInternalServerError()
	}

	// collect allocated names from all accessible namespaces
	var out []*rest_model_zrok.Name
	for _, ns := range namespaces {
		names, err := str.FindNamesWithShareTokensForAccountAndNamespace(int(principal.ID), ns.Id, trx)
		if err != nil {
			dl.Errorf("error finding allocated names for namespace '%v': %v", ns.Token, err)
			return share.NewListAllNamesInternalServerError()
		}

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
	}

	dl.Debugf("listed %d allocated names across all namespaces for account '%v'", len(out), principal.Email)
	return share.NewListAllNamesOK().WithPayload(out)
}
