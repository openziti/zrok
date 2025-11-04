package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/admin"
)

type removeNamespaceFrontendMappingHandler struct{}

func newRemoveNamespaceFrontendMappingHandler() *removeNamespaceFrontendMappingHandler {
	return &removeNamespaceFrontendMappingHandler{}
}

func (handler *removeNamespaceFrontendMappingHandler) Handle(params admin.RemoveNamespaceFrontendMappingParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		dl.Errorf("invalid admin principal")
		return admin.NewRemoveNamespaceFrontendMappingUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return admin.NewRemoveNamespaceFrontendMappingInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	nsToken := params.Body.NamespaceToken
	feToken := params.Body.FrontendToken

	ns, err := str.FindNamespaceWithToken(nsToken, trx)
	if err != nil {
		dl.Errorf("error finding namespace by token '%s': %v", nsToken, err)
		return admin.NewRemoveNamespaceFrontendMappingNotFound()
	}

	fe, err := str.FindFrontendWithToken(feToken, trx)
	if err != nil {
		dl.Errorf("error finding frontend by token '%s': %v", feToken, err)
		return admin.NewRemoveNamespaceFrontendMappingNotFound()
	}

	err = str.DeleteNamespaceFrontendMapping(ns.Id, fe.Id, trx)
	if err != nil {
		dl.Errorf("error deleting namespace frontend mapping: %v", err)
		return admin.NewRemoveNamespaceFrontendMappingInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		dl.Errorf("error committing transaction: %v", err)
		return admin.NewRemoveNamespaceFrontendMappingInternalServerError()
	}

	dl.Infof("removed namespace frontend mapping for namespace '%s' and frontend '%s'", nsToken, feToken)
	return admin.NewRemoveNamespaceFrontendMappingOK()
}
