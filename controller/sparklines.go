package controller

import (
	"slices"

	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/controller/config"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/metadata"
)

type sparklinesHandler struct {
	cfg *config.Config
}

func newSparklinesHandler(cfg *config.Config) *sparklinesHandler {
	return &sparklinesHandler{cfg: cfg}
}

func (h *sparklinesHandler) Handle(params metadata.GetSparklinesParams, principal *rest_model_zrok.Principal) middleware.Responder {
	out := &metadata.GetSparklinesOKBody{}

	if h.cfg.Metrics != nil && h.cfg.Metrics.Influx != nil {
		trx, err := str.Begin()
		if err != nil {
			dl.Errorf("error beginning transaction: %v", err)
			return metadata.NewGetSparklinesInternalServerError()
		}
		defer func() { _ = trx.Rollback() }()

		if len(params.Body.Environments) > 0 {
			if envs, err := str.FindEnvironmentsForAccount(int(principal.ID), trx); err == nil {
				var selectedEnvs []*store.Environment
				selectedEnvsIdIdx := make(map[int]*store.Environment)
				for _, envZId := range params.Body.Environments {
					if idx := slices.IndexFunc(envs, func(env *store.Environment) bool { return env.ZId == envZId }); idx > -1 {
						selectedEnvs = append(selectedEnvs, envs[idx])
						selectedEnvsIdIdx[envs[idx].Id] = envs[idx]
					} else {
						dl.Warnf("requested sparkdata for environment '%v' not owned by '%v'", envZId, principal.Email)
					}
				}
				envsRxSparkdata, envsTxSparkdata, err := sparkDataForEnvironments(selectedEnvs)
				if err != nil {
					dl.Errorf("error getting sparkdata for selected environments for '%v': %v", principal.Email, err)
					return metadata.NewGetSparklinesInternalServerError()
				}
				for envId, rx := range envsRxSparkdata {
					tx := envsTxSparkdata[envId]
					forEnv := selectedEnvsIdIdx[envId]

					var samples []*rest_model_zrok.MetricsSample
					for i := 0; i < len(rx) && i < len(tx); i++ {
						samples = append(samples, &rest_model_zrok.MetricsSample{
							Rx: float64(rx[i]),
							Tx: float64(tx[i]),
						})
					}
					out.Sparklines = append(out.Sparklines, &rest_model_zrok.Metrics{
						Scope:   "environment",
						ID:      forEnv.ZId,
						Samples: samples,
					})
				}
			} else {
				dl.Errorf("error finding environments for '%v': %v", principal.Email, err)
				return metadata.NewGetSparklinesInternalServerError()
			}
		}

		if len(params.Body.Shares) > 0 {
			if shrs, err := str.FindAllSharesForAccount(int(principal.ID), trx); err == nil {
				var selectedShares []*store.Share
				for _, selectedShareToken := range params.Body.Shares {
					if idx := slices.IndexFunc(shrs, func(shr *store.Share) bool { return shr.Token == selectedShareToken }); idx > -1 {
						selectedShares = append(selectedShares, shrs[idx])
					} else {
						dl.Warnf("requested sparkdata for share '%v' not owned by '%v'", selectedShareToken, principal.Email)
					}
				}
				shrsRxSparkdata, shrsTxSparkdata, err := sparkDataForShares(selectedShares)
				if err != nil {
					dl.Errorf("error getting sparkdata for selected shares for '%v': %v", principal.Email, err)
					return metadata.NewGetSparklinesInternalServerError()
				}
				for shrToken, rx := range shrsRxSparkdata {
					tx := shrsTxSparkdata[shrToken]

					var samples []*rest_model_zrok.MetricsSample
					for i := 0; i < len(rx) && i < len(tx); i++ {
						samples = append(samples, &rest_model_zrok.MetricsSample{
							Rx: float64(rx[i]),
							Tx: float64(tx[i]),
						})
					}
					out.Sparklines = append(out.Sparklines, &rest_model_zrok.Metrics{
						Scope:   "share",
						ID:      shrToken,
						Samples: samples,
					})
				}
			} else {
				dl.Errorf("error finding shares for '%v': %v", principal.Email, err)
				return metadata.NewGetSparklinesInternalServerError()
			}
		}
	}

	return metadata.NewGetSparklinesOK().WithPayload(out)
}
