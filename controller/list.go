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
	envs, err := str.FindEnvironmentsForAccount(int(principal.ID), tx)
	if err != nil {
		logrus.Errorf("error finding identities for '%v': %v", principal.Username, err)
		return metadata.NewListEnvironmentsInternalServerError().WithPayload(rest_model_zrok.ErrorMessage(err.Error()))
	}
	var out rest_model_zrok.Environments
	for _, env := range envs {
		out = append(out, &rest_model_zrok.Environment{
			Active:    env.Active,
			CreatedAt: env.CreatedAt.String(),
			UpdatedAt: env.UpdatedAt.String(),
			ZitiID:    env.ZitiIdentityId,
		})
	}
	return metadata.NewListEnvironmentsOK().WithPayload(out)
}
