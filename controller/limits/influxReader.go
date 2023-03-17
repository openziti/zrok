package limits

import (
	"context"
	"fmt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/openziti/zrok/controller/metrics"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

type influxReader struct {
	cfg      *metrics.InfluxConfig
	idb      influxdb2.Client
	queryApi api.QueryAPI
}

func newInfluxReader(cfg *metrics.InfluxConfig) *influxReader {
	idb := influxdb2.NewClient(cfg.Url, cfg.Token)
	queryApi := idb.QueryAPI(cfg.Org)
	return &influxReader{cfg, idb, queryApi}
}

func (r *influxReader) totalRxForAccount(acctId int64, duration time.Duration) (int64, error) {
	query := fmt.Sprintf("from(bucket: \"%v\")\n", r.cfg.Bucket) +
		fmt.Sprintf("|> range(start: -%v)\n", duration) +
		"|> filter(fn: (r) => r[\"_measurement\"] == \"xfer\")\n" +
		"|> filter(fn: (r) => r[\"_field\"] == \"bytesRead\")\n" +
		"|> filter(fn: (r) => r[\"namespace\"] == \"backend\")\n" +
		fmt.Sprintf("|> filter(fn: (r) => r[\"acctId\"] == \"%d\")\n", acctId) +
		"|> sum()"
	return r.runQueryForSum(query)
}

func (r *influxReader) totalTxForAccount(acctId int64, duration time.Duration) (int64, error) {
	query := fmt.Sprintf("from(bucket: \"%v\")\n", r.cfg.Bucket) +
		fmt.Sprintf("|> range(start: -%v)\n", duration) +
		"|> filter(fn: (r) => r[\"_measurement\"] == \"xfer\")\n" +
		"|> filter(fn: (r) => r[\"_field\"] == \"bytesWritten\")\n" +
		"|> filter(fn: (r) => r[\"namespace\"] == \"backend\")\n" +
		fmt.Sprintf("|> filter(fn: (r) => r[\"acctId\"] == \"%d\")\n", acctId) +
		"|> sum()"
	return r.runQueryForSum(query)
}

func (r *influxReader) totalRxForEnvironment(envId int64, duration time.Duration) (int64, error) {
	query := fmt.Sprintf("from(bucket: \"%v\")\n", r.cfg.Bucket) +
		fmt.Sprintf("|> range(start: -%v)\n", duration) +
		"|> filter(fn: (r) => r[\"_measurement\"] == \"xfer\")\n" +
		"|> filter(fn: (r) => r[\"_field\"] == \"bytesRead\")\n" +
		"|> filter(fn: (r) => r[\"namespace\"] == \"backend\")\n" +
		fmt.Sprintf("|> filter(fn: (r) => r[\"envId\"] == \"%d\")\n", envId) +
		"|> sum()"
	return r.runQueryForSum(query)
}

func (r *influxReader) totalTxForEnvironment(envId int64, duration time.Duration) (int64, error) {
	query := fmt.Sprintf("from(bucket: \"%v\")\n", r.cfg.Bucket) +
		fmt.Sprintf("|> range(start: -%v)\n", duration) +
		"|> filter(fn: (r) => r[\"_measurement\"] == \"xfer\")\n" +
		"|> filter(fn: (r) => r[\"_field\"] == \"bytesWritten\")\n" +
		"|> filter(fn: (r) => r[\"namespace\"] == \"backend\")\n" +
		fmt.Sprintf("|> filter(fn: (r) => r[\"envId\"] == \"%d\")\n", envId) +
		"|> sum()"
	return r.runQueryForSum(query)
}

func (r *influxReader) totalRxForShare(shrToken string, duration time.Duration) (int64, error) {
	query := fmt.Sprintf("from(bucket: \"%v\")\n", r.cfg.Bucket) +
		fmt.Sprintf("|> range(start: -%v)\n", duration) +
		"|> filter(fn: (r) => r[\"_measurement\"] == \"xfer\")\n" +
		"|> filter(fn: (r) => r[\"_field\"] == \"bytesRead\")\n" +
		"|> filter(fn: (r) => r[\"namespace\"] == \"backend\")\n" +
		fmt.Sprintf("|> filter(fn: (r) => r[\"share\"] == \"%v\")\n", shrToken) +
		"|> sum()"
	return r.runQueryForSum(query)
}

func (r *influxReader) totalTxForShare(shrToken string, duration time.Duration) (int64, error) {
	query := fmt.Sprintf("from(bucket: \"%v\")\n", r.cfg.Bucket) +
		fmt.Sprintf("|> range(start: -%v)\n", duration) +
		"|> filter(fn: (r) => r[\"_measurement\"] == \"xfer\")\n" +
		"|> filter(fn: (r) => r[\"_field\"] == \"bytesWritten\")\n" +
		"|> filter(fn: (r) => r[\"namespace\"] == \"backend\")\n" +
		fmt.Sprintf("|> filter(fn: (r) => r[\"share\"] == \"%v\")\n", shrToken) +
		"|> sum()"
	return r.runQueryForSum(query)
}

func (r *influxReader) runQueryForSum(query string) (int64, error) {
	result, err := r.queryApi.Query(context.Background(), query)
	if err != nil {
		return -1, err
	}

	if result.Next() {
		if v, ok := result.Record().Value().(int64); ok {
			return v, nil
		} else {
			return -1, errors.New("error asserting result type")
		}
	}

	logrus.Warnf("empty read result set for '%v'", strings.ReplaceAll(query, "\n", ""))
	return 0, nil
}
