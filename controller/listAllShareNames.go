package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/share"
	"github.com/sirupsen/logrus"
)

type listAllShareNamesHandler struct{}

func newListAllShareNamesHandler() *listAllShareNamesHandler {
	return &listAllShareNamesHandler{}
}

func (h *listAllShareNamesHandler) Handle(params share.ListAllShareNamesParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return share.NewListAllShareNamesInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	// find all namespaces the user has access to
	namespaces, err := str.FindNamespacesForAccount(int(principal.ID), trx)
	if err != nil {
		logrus.Errorf("error finding namespaces for account '%v': %v", principal.Email, err)
		return share.NewListAllShareNamesInternalServerError()
	}

	// collect allocated names from all accessible namespaces
	var allNames []*share.ListAllShareNamesOKBodyItems0
	for _, ns := range namespaces {
		allocatedNames, err := str.FindAllocatedNamesForAccountAndNamespace(int(principal.ID), ns.Id, trx)
		if err != nil {
			logrus.Errorf("error finding allocated names for namespace '%v': %v", ns.Token, err)
			return share.NewListAllShareNamesInternalServerError()
		}

		for _, an := range allocatedNames {
			nameObj := &share.ListAllShareNamesOKBodyItems0{
				Name:           an.Name,
				CreatedAt:      an.CreatedAt.Unix(),
				NamespaceName:  ns.Name,
				NamespaceToken: ns.Token,
			}
			allNames = append(allNames, nameObj)
		}
	}

	logrus.Debugf("listed %d allocated names across all namespaces for account '%v'", len(allNames), principal.Email)
	return share.NewListAllShareNamesOK().WithPayload(allNames)
}