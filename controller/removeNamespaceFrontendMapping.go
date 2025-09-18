package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/admin"
	"github.com/sirupsen/logrus"
)

type removeNamespaceFrontendMappingHandler struct{}

func newRemoveNamespaceFrontendMappingHandler() *removeNamespaceFrontendMappingHandler {
	return &removeNamespaceFrontendMappingHandler{}
}

func (handler *removeNamespaceFrontendMappingHandler) Handle(params admin.RemoveNamespaceFrontendMappingParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		logrus.Errorf("invalid admin principal")
		return admin.NewRemoveNamespaceFrontendMappingUnauthorized()
	}

	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return admin.NewRemoveNamespaceFrontendMappingInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()

	nsToken := params.Body.NamespaceToken
	feToken := params.Body.FrontendToken

	ns, err := str.FindNamespaceWithToken(nsToken, trx)
	if err != nil {
		logrus.Errorf("error finding namespace by token '%s': %v", nsToken, err)
		return admin.NewRemoveNamespaceFrontendMappingNotFound()
	}

	fe, err := str.FindFrontendWithToken(feToken, trx)
	if err != nil {
		logrus.Errorf("error finding frontend by token '%s': %v", feToken, err)
		return admin.NewRemoveNamespaceFrontendMappingNotFound()
	}

	err = str.DeleteNamespaceFrontendMapping(ns.Id, fe.Id, trx)
	if err != nil {
		logrus.Errorf("error deleting namespace frontend mapping: %v", err)
		return admin.NewRemoveNamespaceFrontendMappingInternalServerError()
	}

	if err := trx.Commit(); err != nil {
		logrus.Errorf("error committing transaction: %v", err)
		return admin.NewRemoveNamespaceFrontendMappingInternalServerError()
	}

	logrus.Infof("removed namespace frontend mapping for namespace '%s' and frontend '%s'", nsToken, feToken)
	return admin.NewRemoveNamespaceFrontendMappingOK()
}
