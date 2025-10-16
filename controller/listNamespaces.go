package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/admin"
)

type listNamespacesHandler struct{}

func newListNamespacesHandler() *listNamespacesHandler {
	return &listNamespacesHandler{}
}

func (h *listNamespacesHandler) Handle(_ admin.ListNamespacesParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		dl.Error("invalid admin principal")
		return admin.NewListNamespacesUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return admin.NewListNamespacesInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	namespaces, err := str.FindNamespaces(trx)
	if err != nil {
		dl.Errorf("error finding namespaces: %v", err)
		return admin.NewListNamespacesInternalServerError()
	}

	var out []*admin.ListNamespacesOKBodyItems0
	for _, ns := range namespaces {
		out = append(out, &admin.ListNamespacesOKBodyItems0{
			NamespaceToken: ns.Token,
			Name:           ns.Name,
			Description:    ns.Description,
			Open:           ns.Open,
			CreatedAt:      ns.CreatedAt.Unix(),
			UpdatedAt:      ns.UpdatedAt.Unix(),
		})
	}
	return admin.NewListNamespacesOK().WithPayload(out)
}
