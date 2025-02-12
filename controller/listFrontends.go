package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/admin"
	"github.com/sirupsen/logrus"
)

type listFrontendsHandler struct{}

func newListFrontendsHandler() *listFrontendsHandler {
	return &listFrontendsHandler{}
}

func (h *listFrontendsHandler) Handle(params admin.ListFrontendsParams, principal *rest_model_zrok.Principal) middleware.Responder {
	if !principal.Admin {
		logrus.Error("invalid admin principal")
		return admin.NewListFrontendsUnauthorized()
	}

	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return admin.NewListFrontendsInternalServerError()
	}
	defer func() { _ = tx.Rollback() }()

	sfes, err := str.FindPublicFrontends(tx)
	if err != nil {
		logrus.Errorf("error finding public frontends: %v", err)
		return admin.NewListFrontendsInternalServerError()
	}

	var frontends []*admin.ListFrontendsOKBodyItems0
	for _, sfe := range sfes {
		frontend := &admin.ListFrontendsOKBodyItems0{
			FrontendToken: sfe.Token,
			ZID:           sfe.ZId,
			CreatedAt:     sfe.CreatedAt.UnixMilli(),
			UpdatedAt:     sfe.UpdatedAt.UnixMilli(),
		}
		if sfe.UrlTemplate != nil {
			frontend.URLTemplate = *sfe.UrlTemplate
		}
		if sfe.PublicName != nil {
			frontend.PublicName = *sfe.PublicName
		}
		frontends = append(frontends, frontend)
	}
	return admin.NewListFrontendsOK().WithPayload(frontends)
}
