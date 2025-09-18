package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/metadata"
	"github.com/sirupsen/logrus"
)

type accountDetailHandler struct{}

func newAccountDetailHandler() *accountDetailHandler {
	return &accountDetailHandler{}
}

func (h *accountDetailHandler) Handle(params metadata.GetAccountDetailParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction for '%v': %v", principal.Email, err)
		return metadata.NewGetAccountDetailInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()
	envs, err := str.FindEnvironmentsForAccount(int(principal.ID), trx)
	if err != nil {
		logrus.Errorf("error retrieving environments for '%v': %v", principal.Email, err)
		return metadata.NewGetAccountDetailInternalServerError()
	}
	sparkRx := make(map[int][]int64)
	sparkTx := make(map[int][]int64)
	if cfg.Metrics != nil && cfg.Metrics.Influx != nil {
		sparkRx, sparkTx, err = sparkDataForEnvironments(envs)
		if err != nil {
			logrus.Errorf("error querying spark data for environments for '%v': %v", principal.Email, err)
		}
	} else {
		logrus.Debug("skipping spark data for environments; no influx configuration")
	}
	var payload []*rest_model_zrok.Environment
	for _, env := range envs {
		var sparkData []*rest_model_zrok.SparkDataSample
		for i := 0; i < len(sparkRx[env.Id]) && i < len(sparkTx[env.Id]); i++ {
			sparkData = append(sparkData, &rest_model_zrok.SparkDataSample{Rx: float64(sparkRx[env.Id][i]), Tx: float64(sparkTx[env.Id][i])})
		}
		payload = append(payload, &rest_model_zrok.Environment{
			Activity:    sparkData,
			Address:     env.Address,
			CreatedAt:   env.CreatedAt.UnixMilli(),
			Description: env.Description,
			Host:        env.Host,
			UpdatedAt:   env.UpdatedAt.UnixMilli(),
			ZID:         env.ZId,
		})
	}
	return metadata.NewGetAccountDetailOK().WithPayload(payload)
}
