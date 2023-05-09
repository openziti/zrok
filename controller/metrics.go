package controller

import (
	"context"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/openziti/zrok/controller/metrics"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/metadata"
	"github.com/sirupsen/logrus"
	"time"
)

type getAccountMetricsHandler struct {
	cfg      *metrics.InfluxConfig
	idb      influxdb2.Client
	queryApi api.QueryAPI
}

func newGetAccountMetricsHandler(cfg *metrics.InfluxConfig) *getAccountMetricsHandler {
	idb := influxdb2.NewClient(cfg.Url, cfg.Token)
	queryApi := idb.QueryAPI(cfg.Org)
	return &getAccountMetricsHandler{
		cfg:      cfg,
		idb:      idb,
		queryApi: queryApi,
	}
}

func (h *getAccountMetricsHandler) Handle(params metadata.GetAccountMetricsParams, principal *rest_model_zrok.Principal) middleware.Responder {
	duration := 30 * 24 * time.Hour
	if params.Duration != nil {
		v, err := time.ParseDuration(*params.Duration)
		if err != nil {
			logrus.Errorf("bad duration '%v' for '%v': %v", params.Duration, principal.Email, err)
			return metadata.NewGetAccountMetricsBadRequest()
		}
		duration = v
	}
	slice := duration / 200

	query := fmt.Sprintf("from(bucket: \"%v\")\n", h.cfg.Bucket) +
		fmt.Sprintf("|> range(start -%v)\n", duration) +
		"|> filter(fn: (r) => r[\"_measurement\"] == \"xfer\")\n" +
		"|> filter(fn: (r) => r[\"_field\"] == \"rx\" or r[\"_field\"] == \"tx\")\n" +
		"|> filter(fn: (r) => r[\"namespace\"] == \"backend\")\n" +
		fmt.Sprintf("|> filter(fn: (r) => r[\"acctId\"] == \"%d\")\n", principal.ID) +
		"|> drop(columns: [\"share\", \"envId\"])\n" +
		fmt.Sprintf("|> aggregateWindow(every: %v, fn: sum, createEmpty: true)", slice)

	rx, tx, err := runFluxForRxTxArray(query, h.queryApi)
	if err != nil {
		logrus.Errorf("error running account metrics query for '%v': %v", principal.Email, err)
		return metadata.NewGetAccountMetricsInternalServerError()
	}

	response := &rest_model_zrok.Metrics{
		ID:     fmt.Sprintf("%d", principal.ID),
		Period: duration.Seconds(),
		Rx:     rx,
		Tx:     tx,
	}
	return metadata.NewGetAccountMetricsOK().WithPayload(response)
}

func runFluxForRxTxArray(query string, queryApi api.QueryAPI) (rx, tx []float64, err error) {
	result, err := queryApi.Query(context.Background(), query)
	if err != nil {
		return nil, nil, err
	}
	for result.Next() {
		if v, ok := result.Record().Value().(int64); ok {
			switch result.Record().Field() {
			case "rx":
				rx = append(rx, float64(v))
			case "tx":
				tx = append(tx, float64(v))
			}
		}
	}
	return rx, tx, nil
}
