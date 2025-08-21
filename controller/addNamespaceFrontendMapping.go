package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/admin"
	"github.com/sirupsen/logrus"
)

type addNamespaceFrontendMappingHandler struct{}

func newAddNamespaceFrontendMappingHandler() *addNamespaceFrontendMappingHandler {
	return &addNamespaceFrontendMappingHandler{}
}

func (handler *addNamespaceFrontendMappingHandler) Handle(params admin.AddNamespaceFrontendMappingParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		logrus.Errorf("invalid admin principal")
		return admin.NewAddNamespaceFrontendMappingUnauthorized()
	}

	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return admin.NewAddNamespaceFrontendMappingInternalServerError()
	}
	defer func() { _ = tx.Rollback() }()

	nsToken := params.Body.NamespaceToken
	feToken := params.Body.FrontendToken

	ns, err := str.FindNamespaceWithToken(nsToken, tx)
	if err != nil {
		logrus.Errorf("error finding namespace by token '%s': %v", nsToken, err)
		return admin.NewAddNamespaceFrontendMappingNotFound()
	}

	fe, err := str.FindFrontendWithToken(feToken, tx)
	if err != nil {
		logrus.Errorf("error finding frontend by token '%s': %v", feToken, err)
		return admin.NewAddNamespaceFrontendMappingNotFound()
	}

	_, err = str.CreateNamespaceFrontendMapping(ns.Id, fe.Id, tx)
	if err != nil {
		logrus.Errorf("error creating namespace frontend mapping: %v", err)
		return admin.NewAddNamespaceFrontendMappingInternalServerError()
	}

	if err := tx.Commit(); err != nil {
		logrus.Errorf("error committing transaction: %v", err)
		return admin.NewAddNamespaceFrontendMappingInternalServerError()
	}

	logrus.Infof("added namespace frontend mapping for namespace '%s' and frontend '%s'", nsToken, feToken)
	return admin.NewAddNamespaceFrontendMappingOK()
}
