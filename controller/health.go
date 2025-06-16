package controller

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

func HealthCheckHTTP(w http.ResponseWriter, _ *http.Request) {
	if err := healthCheckStore(w); err != nil {
		logrus.Error(err)
		return
	}
	if err := healthCheckMetrics(w); err != nil {
		logrus.Error(err)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte("<html><body><h1>Healthy</h1></body></html>"))
}

func healthCheckStore(w http.ResponseWriter) error {
	trx, err := str.Begin()
	if err != nil {
		http.Error(w, "error starting transaction", http.StatusInternalServerError)
		return err
	}
	defer func() {
		_ = trx.Rollback()
	}()
	count := -1
	if err := trx.QueryRowx("select count(0) from migrations").Scan(&count); err != nil {
		http.Error(w, "error selecting migration count", http.StatusInternalServerError)
		return err
	}
	logrus.Debugf("%d migrations", count)
	return nil
}

func healthCheckMetrics(w http.ResponseWriter) error {
	if cfg.Metrics != nil && cfg.Metrics.Influx != nil {
		queryApi := idb.QueryAPI(cfg.Metrics.Influx.Org)
		query := fmt.Sprintf("from(bucket: \"%v\")\n", cfg.Metrics.Influx.Bucket) +
			"|> range(start: -5s)\n" +
			"|> filter(fn: (r) => r[\"_measurement\"] == \"xfer\")\n" +
			"|> filter(fn: (r) => r[\"_field\"] == \"rx\" or r[\"_field\"] == \"tx\")\n" +
			"|> filter(fn: (r) => r[\"namespace\"] == \"backend\")\n" +
			"|> sum()"
		result, err := queryApi.Query(context.Background(), query)
		if err != nil {
			http.Error(w, "error querying influx", http.StatusInternalServerError)
			return err
		}
		results := 0
		for result.Next() {
			results++
		}
		logrus.Debugf("%d results", results)
	}
	return nil
}
