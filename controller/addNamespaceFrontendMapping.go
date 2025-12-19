package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/admin"
)

type addNamespaceFrontendMappingHandler struct{}

func newAddNamespaceFrontendMappingHandler() *addNamespaceFrontendMappingHandler {
	return &addNamespaceFrontendMappingHandler{}
}

func (handler *addNamespaceFrontendMappingHandler) Handle(params admin.AddNamespaceFrontendMappingParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		dl.Errorf("invalid admin principal")
		return admin.NewAddNamespaceFrontendMappingUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return admin.NewAddNamespaceFrontendMappingInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	nsToken := params.Body.NamespaceToken
	feToken := params.Body.FrontendToken
	isDefault := params.Body.IsDefault

	ns, err := str.FindNamespaceWithToken(nsToken, trx)
	if err != nil {
		dl.Errorf("error finding namespace by token '%s': %v", nsToken, err)
		return admin.NewAddNamespaceFrontendMappingNotFound()
	}

	fe, err := str.FindFrontendWithToken(feToken, trx)
	if err != nil {
		dl.Errorf("error finding frontend by token '%s': %v", feToken, err)
		return admin.NewAddNamespaceFrontendMappingNotFound()
	}

	_, err = str.CreateNamespaceFrontendMapping(ns.Id, fe.Id, isDefault, trx)
	if err != nil {
		dl.Errorf("error creating namespace frontend mapping: %v", err)
		return admin.NewAddNamespaceFrontendMappingInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		dl.Errorf("error committing transaction: %v", err)
		return admin.NewAddNamespaceFrontendMappingInternalServerError()
	}

	dl.Infof("added namespace frontend mapping for namespace '%s' and frontend '%s'", nsToken, feToken)
	return admin.NewAddNamespaceFrontendMappingOK()
}
