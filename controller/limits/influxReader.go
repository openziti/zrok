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

func (r *influxReader) totalRxTxForAccount(acctId int64, duration time.Duration) (int64, int64, error) {
	query := fmt.Sprintf("from(bucket: \"%v\")\n", r.cfg.Bucket) +
		fmt.Sprintf("|> range(start: -%v)\n", duration) +
		"|> filter(fn: (r) => r[\"_measurement\"] == \"xfer\")\n" +
		"|> filter(fn: (r) => r[\"_field\"] == \"rx\" or r[\"_field\"] == \"tx\")\n" +
		"|> filter(fn: (r) => r[\"namespace\"] == \"backend\")\n" +
		fmt.Sprintf("|> filter(fn: (r) => r[\"acctId\"] == \"%d\")\n", acctId) +
		"|> drop(columns: [\"share\", \"envId\"])\n" +
		"|> sum()"
	return r.runQueryForRxTx(query)
}

func (r *influxReader) totalRxTxForEnvironment(envId int64, duration time.Duration) (int64, int64, error) {
	query := fmt.Sprintf("from(bucket: \"%v\")\n", r.cfg.Bucket) +
		fmt.Sprintf("|> range(start: -%v)\n", duration) +
		"|> filter(fn: (r) => r[\"_measurement\"] == \"xfer\")\n" +
		"|> filter(fn: (r) => r[\"_field\"] == \"rx\" or r[\"_field\"] == \"tx\")\n" +
		"|> filter(fn: (r) => r[\"namespace\"] == \"backend\")\n" +
		fmt.Sprintf("|> filter(fn: (r) => r[\"envId\"] == \"%d\")\n", envId) +
		"|> drop(columns: [\"share\", \"acctId\"])\n" +
		"|> sum()"
	return r.runQueryForRxTx(query)
}

func (r *influxReader) totalRxTxForShare(shrToken string, duration time.Duration) (int64, int64, error) {
	query := fmt.Sprintf("from(bucket: \"%v\")\n", r.cfg.Bucket) +
		fmt.Sprintf("|> range(start: -%v)\n", duration) +
		"|> filter(fn: (r) => r[\"_measurement\"] == \"xfer\")\n" +
		"|> filter(fn: (r) => r[\"_field\"] == \"rx\" or r[\"_field\"] == \"tx\")\n" +
		"|> filter(fn: (r) => r[\"namespace\"] == \"backend\")\n" +
		fmt.Sprintf("|> filter(fn: (r) => r[\"share\"] == \"%v\")\n", shrToken) +
		"|> sum()"
	return r.runQueryForRxTx(query)
}

func (r *influxReader) runQueryForRxTx(query string) (rx int64, tx int64, err error) {
	result, err := r.queryApi.Query(context.Background(), query)
	if err != nil {
		return -1, -1, err
	}

	count := 0
	for result.Next() {
		if v, ok := result.Record().Value().(int64); ok {
			switch result.Record().Field() {
			case "tx":
				tx = v
			case "rx":
				rx = v
			default:
				logrus.Warnf("field '%v'?", result.Record().Field())
			}
		} else {
			return -1, -1, errors.New("error asserting value type")
		}
		count++
	}
	if count != 0 && count != 2 {
		return -1, -1, errors.Errorf("expected 2 results; got '%d' (%v)", count, strings.ReplaceAll(query, "\n", ""))
	}
	return rx, tx, nil
}
