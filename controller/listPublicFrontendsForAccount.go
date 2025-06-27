package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/metadata"
	"github.com/sirupsen/logrus"
)

type listPublicFrontendsForAccountHandler struct{}

func newListPublicFrontendsForAccountHandler() *listPublicFrontendsForAccountHandler {
	return &listPublicFrontendsForAccountHandler{}
}

func (h *listPublicFrontendsForAccountHandler) Handle(_ metadata.ListPublicFrontendsForAccountParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return metadata.NewListPublicFrontendsForAccountInternalServerError()
	}
	defer trx.Rollback()

	var publicFrontends []*metadata.ListPublicFrontendsForAccountOKBodyPublicFrontendsItems0

	openFes, err := str.FindOpenPublicFrontends(trx)
	if err != nil {
		logrus.Errorf("error finding open public frontends: %v", err)
		return metadata.NewListPublicFrontendsForAccountInternalServerError()
	}
	for _, openFe := range openFes {
		publicFrontends = append(publicFrontends, &metadata.ListPublicFrontendsForAccountOKBodyPublicFrontendsItems0{
			PublicName:  *openFe.PublicName,
			URLTemplate: *openFe.UrlTemplate,
		})
	}

	closedFes, err := str.FindClosedPublicFrontendsGrantedToAccount(int(principal.ID), trx)
	if err != nil {
		logrus.Errorf("error finding closed public frontends: %v", err)
		return metadata.NewListPublicFrontendsForAccountInternalServerError()
	}
	for _, closedFe := range closedFes {
		publicFrontends = append(publicFrontends, &metadata.ListPublicFrontendsForAccountOKBodyPublicFrontendsItems0{
			PublicName:  *closedFe.PublicName,
			URLTemplate: *closedFe.UrlTemplate,
		})
	}

	payload := &metadata.ListPublicFrontendsForAccountOKBody{PublicFrontends: publicFrontends}
	return metadata.NewListPublicFrontendsForAccountOK().WithPayload(payload)
}
