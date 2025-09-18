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

	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return admin.NewListNamespaceFrontendMappingsInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	nsToken := params.NamespaceToken

	ns, err := str.FindNamespaceWithToken(nsToken, trx)
	if err != nil {
		logrus.Errorf("error finding namespace by token '%s': %v", nsToken, err)
		return admin.NewListNamespaceFrontendMappingsNotFound()
	}

	nfMappings, err := str.FindNamespaceFrontendMappingsForNamespace(ns.Id, trx)
	if err != nil {
		logrus.Errorf("error finding namespace frontend mappings for namespace '%s': %v", nsToken, err)
		return admin.NewListNamespaceFrontendMappingsInternalServerError()
	}

	var mappings []*admin.ListNamespaceFrontendMappingsOKBodyItems0
	for _, nfMapping := range nfMappings {
		fe, err := str.GetFrontend(nfMapping.FrontendId, trx)
		if err != nil {
			logrus.Errorf("error finding frontend with id '%d': %v", nfMapping.FrontendId, err)
			continue
		}
		mapping := &admin.ListNamespaceFrontendMappingsOKBodyItems0{
			NamespaceToken: nsToken,
			FrontendToken:  fe.Token,
			IsDefault:      nfMapping.IsDefault,
			CreatedAt:      nfMapping.CreatedAt.Unix(),
		}
		mappings = append(mappings, mapping)
	}

	logrus.Infof("listed '%d' frontend mappings for namespace '%s'", len(mappings), nsToken)
	return admin.NewListNamespaceFrontendMappingsOK().WithPayload(mappings)
}
