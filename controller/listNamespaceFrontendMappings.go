package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/admin"
	"github.com/sirupsen/logrus"
)

type listNamespaceFrontendMappingsHandler struct{}

func newListNamespaceFrontendMappingsHandler() *listNamespaceFrontendMappingsHandler {
	return &listNamespaceFrontendMappingsHandler{}
}

func (handler *listNamespaceFrontendMappingsHandler) Handle(params admin.ListNamespaceFrontendMappingsParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		logrus.Errorf("invalid admin principal")
		return admin.NewListNamespaceFrontendMappingsUnauthorized()
	}

	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return admin.NewListNamespaceFrontendMappingsInternalServerError()
	}
	defer func() { _ = tx.Rollback() }()

	nsToken := params.NamespaceToken

	ns, err := str.FindNamespaceWithToken(nsToken, tx)
	if err != nil {
		logrus.Errorf("error finding namespace by token '%s': %v", nsToken, err)
		return admin.NewListNamespaceFrontendMappingsNotFound()
	}

	frontends, err := str.FindFrontendsForNamespace(ns.Id, tx)
	if err != nil {
		logrus.Errorf("error finding frontends for namespace '%s': %v", nsToken, err)
		return admin.NewListNamespaceFrontendMappingsInternalServerError()
	}

	var mappings []*admin.ListNamespaceFrontendMappingsOKBodyItems0
	for _, fe := range frontends {
		mapping := &admin.ListNamespaceFrontendMappingsOKBodyItems0{
			NamespaceToken: nsToken,
			FrontendToken:  fe.Token,
			CreatedAt:      fe.CreatedAt.Unix(),
		}
		mappings = append(mappings, mapping)
	}

	logrus.Infof("listed '%d' frontend mappings for namespace '%s'", len(mappings), nsToken)
	return admin.NewListNamespaceFrontendMappingsOK().WithPayload(mappings)
}
