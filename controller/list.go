package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/metadata"
	"github.com/sirupsen/logrus"
)

func listEnvironmentsHandler(_ metadata.ListEnvironmentsParams, principal *rest_model_zrok.Principal) middleware.Responder {
	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return metadata.NewListEnvironmentsInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	defer func() { _ = tx.Rollback() }()
	ids, err := str.FindIdentitiesForAccount(int(principal.ID), tx)
	if err != nil {
		logrus.Errorf("error finding identities for '%v': %v", principal.Username, err)
		return metadata.NewListEnvironmentsInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	var out rest_model_zrok.Environments
	for _, id := range ids {
		out = append(out, &rest_model_zrok.Environment{
			Active:    id.Active,
			CreatedAt: id.CreatedAt.String(),
			UpdatedAt: id.UpdatedAt.String(),
			ZitiID:    id.ZitiId,
		})
	}
	return metadata.NewListEnvironmentsOK().WithPayload(out)
}
