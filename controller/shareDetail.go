package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/metadata"
)

type shareDetailHandler struct{}

func newShareDetailHandler() *shareDetailHandler {
	return &shareDetailHandler{}
}

func (h *shareDetailHandler) Handle(params metadata.GetShareDetailParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction: %v", err)
		return metadata.NewGetShareDetailInternalServerError()
	}
	defer func() { _ = trx.Rollback() }()
	shr, err := str.FindShareWithToken(params.ShareToken, trx)
	if err != nil {
		dl.Errorf("error finding share '%v': %v", params.ShareToken, err)
		return metadata.NewGetShareDetailNotFound()
	}
	envs, err := str.FindEnvironmentsForAccount(int(principal.ID), trx)
	if err != nil {
		dl.Errorf("error finding environments for account '%v': %v", principal.Email, err)
		return metadata.NewGetShareDetailInternalServerError()
	}
	found := false
	var shrEnv *store.Environment
	for _, env := range envs {
		if shr.EnvironmentId == env.Id {
			shrEnv = env
			found = true
			break
		}
	}
	if !found {
		dl.Errorf("environment not matched for share '%v' for account '%v'", params.ShareToken, principal.Email)
		return metadata.NewGetShareDetailNotFound()
	}
	sparkRx := make(map[string][]int64)
	sparkTx := make(map[string][]int64)
	if cfg.Metrics != nil && cfg.Metrics.Influx != nil {
		sparkRx, sparkTx, err = sparkDataForShares([]*store.Share{shr})
		if err != nil {
			dl.Errorf("error querying spark data for share: %v", err)
		}
	} else {
		dl.Debug("skipping spark data; no influx configuration")
	}

	frontendEndpoints := buildFrontendEndpointsForShare(shr.Id, shr.Token, shr.FrontendEndpoint, trx)

	target := ""
	if shr.BackendProxyEndpoint != nil {
		target = *shr.BackendProxyEndpoint
	}
	var sparkData []*rest_model_zrok.SparkDataSample
	for i := 0; i < len(sparkRx[shr.Token]) && i < len(sparkTx[shr.Token]); i++ {
		sparkData = append(sparkData, &rest_model_zrok.SparkDataSample{Rx: float64(sparkRx[shr.Token][i]), Tx: float64(sparkTx[shr.Token][i])})
	}
	return metadata.NewGetShareDetailOK().WithPayload(&rest_model_zrok.Share{
		ShareToken:        shr.Token,
		ZID:               shr.ZId,
		EnvZID:            shrEnv.ZId,
		ShareMode:         shr.ShareMode,
		BackendMode:       shr.BackendMode,
		FrontendEndpoints: frontendEndpoints,
		Target:            target,
		Activity:          sparkData,
		CreatedAt:         shr.CreatedAt.UnixMilli(),
		UpdatedAt:         shr.UpdatedAt.UnixMilli(),
	})
}
