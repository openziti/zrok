package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/metadata"
	"github.com/sirupsen/logrus"
)

type shareDetailHandler struct{}

func newShareDetailHandler() *shareDetailHandler {
	return &shareDetailHandler{}
}

func (h *shareDetailHandler) Handle(params metadata.GetShareDetailParams, principal *rest_model_zrok.Principal) middleware.Responder {
	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return metadata.NewGetShareDetailInternalServerError()
	}
	defer func() { _ = tx.Rollback() }()
	shr, err := str.FindShareWithToken(params.ShareToken, tx)
	if err != nil {
		logrus.Errorf("error finding share '%v': %v", params.ShareToken, err)
		return metadata.NewGetShareDetailNotFound()
	}
	envs, err := str.FindEnvironmentsForAccount(int(principal.ID), tx)
	if err != nil {
		logrus.Errorf("error finding environments for account '%v': %v", principal.Email, err)
		return metadata.NewGetShareDetailInternalServerError()
	}
	found := false
	for _, env := range envs {
		if shr.EnvironmentId == env.Id {
			found = true
			break
		}
	}
	if !found {
		logrus.Errorf("environment not matched for share '%v' for account '%v'", params.ShareToken, principal.Email)
		return metadata.NewGetShareDetailNotFound()
	}
	sparkRx := make(map[string][]int64)
	sparkTx := make(map[string][]int64)
	if cfg.Metrics != nil && cfg.Metrics.Influx != nil {
		sparkRx, sparkTx, err = sparkDataForShares([]*store.Share{shr})
		if err != nil {
			logrus.Errorf("error querying spark data for share: %v", err)
		}
	} else {
		logrus.Debug("skipping spark data; no influx configuration")
	}
	feEndpoint := ""
	if shr.FrontendEndpoint != nil {
		feEndpoint = *shr.FrontendEndpoint
	}
	feSelection := ""
	if shr.FrontendSelection != nil {
		feSelection = *shr.FrontendSelection
	}
	beProxyEndpoint := ""
	if shr.BackendProxyEndpoint != nil {
		beProxyEndpoint = *shr.BackendProxyEndpoint
	}
	var sparkData []*rest_model_zrok.SparkDataSample
	for i := 0; i < len(sparkRx[shr.Token]) && i < len(sparkTx[shr.Token]); i++ {
		sparkData = append(sparkData, &rest_model_zrok.SparkDataSample{Rx: float64(sparkRx[shr.Token][i]), Tx: float64(sparkTx[shr.Token][i])})
	}
	return metadata.NewGetShareDetailOK().WithPayload(&rest_model_zrok.Share{
		ShareToken:           shr.Token,
		ZID:                  shr.ZId,
		ShareMode:            shr.ShareMode,
		BackendMode:          shr.BackendMode,
		FrontendSelection:    feSelection,
		FrontendEndpoint:     feEndpoint,
		BackendProxyEndpoint: beProxyEndpoint,
		Reserved:             shr.Reserved,
		Activity:             sparkData,
		CreatedAt:            shr.CreatedAt.UnixMilli(),
		UpdatedAt:            shr.UpdatedAt.UnixMilli(),
	})
}
