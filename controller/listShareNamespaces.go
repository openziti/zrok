package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/share"
)

type listShareNamespacesHandler struct{}

func newListShareNamespacesHandler() *listShareNamespacesHandler {
	return &listShareNamespacesHandler{}
}

func (h *listShareNamespacesHandler) Handle(params share.ListShareNamespacesParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction for account '%v': %v", principal.Email, err)
		return share.NewListShareNamespacesInternalServerError()
	}
	defer trx.Rollback()

	// find all namespaces the user has access to
	namespaces, err := str.FindNamespacesForAccount(int(principal.ID), trx)
	if err != nil {
		dl.Errorf("error finding namespaces for account '%v': %v", principal.Email, err)
		return share.NewListShareNamespacesInternalServerError()
	}

	var out []*share.ListShareNamespacesOKBodyItems0
	for _, namespace := range namespaces {
		out = append(out, &share.ListShareNamespacesOKBodyItems0{
			NamespaceToken: namespace.Token,
			Name:           namespace.Name,
			Description:    namespace.Description,
		})
	}

	return share.NewListShareNamespacesOK().WithPayload(out)
}
