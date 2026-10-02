package limits

import (
	"context"
	"fmt"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/metrics"
	"github.com/pkg/errors"
)

type influxReader struct {
	cfg      *metrics.InfluxConfig
	idb      influxdb2.Client
	queryApi api.QueryAPI
	timeout  time.Duration
}

func newInfluxReader(cfg *metrics.InfluxConfig, timeout time.Duration) *influxReader {
	if timeout <= 0 {
		timeout = DefaultQueryTimeout
	}
	idb := influxdb2.NewClient(cfg.Url, cfg.Token)
	queryApi := idb.QueryAPI(cfg.Org)
	return &influxReader{cfg, idb, queryApi, timeout}
}

func (r *influxReader) totalRxTxForAccount(ctx context.Context, acctId int64, duration time.Duration) (int64, int64, error) {
	query := fmt.Sprintf("from(bucket: \"%v\")\n", r.cfg.Bucket) +
		fmt.Sprintf("|> range(start: -%v)\n", duration) +
		"|> filter(fn: (r) => r[\"_measurement\"] == \"xfer\")\n" +
		"|> filter(fn: (r) => r[\"_field\"] == \"rx\" or r[\"_field\"] == \"tx\")\n" +
		"|> filter(fn: (r) => r[\"namespace\"] == \"backend\")\n" +
		fmt.Sprintf("|> filter(fn: (r) => r[\"acctId\"] == \"%d\")\n", acctId) +
		"|> drop(columns: [\"share\", \"envId\"])\n" +
		"|> sum()"
	return r.runQueryForRxTx(ctx, query)
}

func (r *influxReader) totalRxTxForEnvironment(ctx context.Context, envId int64, duration time.Duration) (int64, int64, error) {
	query := fmt.Sprintf("from(bucket: \"%v\")\n", r.cfg.Bucket) +
		fmt.Sprintf("|> range(start: -%v)\n", duration) +
		"|> filter(fn: (r) => r[\"_measurement\"] == \"xfer\")\n" +
		"|> filter(fn: (r) => r[\"_field\"] == \"rx\" or r[\"_field\"] == \"tx\")\n" +
		"|> filter(fn: (r) => r[\"namespace\"] == \"backend\")\n" +
		fmt.Sprintf("|> filter(fn: (r) => r[\"envId\"] == \"%d\")\n", envId) +
		"|> drop(columns: [\"share\", \"acctId\"])\n" +
		"|> sum()"
	return r.runQueryForRxTx(ctx, query)
}

func (r *influxReader) totalRxTxForShare(ctx context.Context, shrToken string, duration time.Duration) (int64, int64, error) {
	query := fmt.Sprintf("from(bucket: \"%v\")\n", r.cfg.Bucket) +
		fmt.Sprintf("|> range(start: -%v)\n", duration) +
		"|> filter(fn: (r) => r[\"_measurement\"] == \"xfer\")\n" +
		"|> filter(fn: (r) => r[\"_field\"] == \"rx\" or r[\"_field\"] == \"tx\")\n" +
		"|> filter(fn: (r) => r[\"namespace\"] == \"backend\")\n" +
		fmt.Sprintf("|> filter(fn: (r) => r[\"share\"] == \"%v\")\n", shrToken) +
		"|> sum()"
	return r.runQueryForRxTx(ctx, query)
}

// runQueryForRxTx bounds the query, reading its result included, by the reader's timeout and by ctx.
func (r *influxReader) runQueryForRxTx(ctx context.Context, query string) (rx int64, tx int64, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	result, err := r.queryApi.Query(ctx, query)
	if err != nil {
		return -1, -1, err
	}
	defer func() { _ = result.Close() }()

	count := 0
	for result.Next() {
		if v, ok := result.Record().Value().(int64); ok {
			switch result.Record().Field() {
			case "tx":
				tx = v
			case "rx":
				rx = v
			default:
				dl.Warnf("field '%v'?", result.Record().Field())
			}
		} else {
			return -1, -1, errors.New("error asserting value type")
		}
		count++
	}
	// a deadline or cancellation part-way through the result ends Next early; reporting that as no usage
	// would relax a limited account.
	if err := result.Err(); err != nil {
		return -1, -1, err
	}
	if count != 0 && count != 2 {
		return -1, -1, errors.Errorf("expected 2 results; got '%d' (%v)", count, strings.ReplaceAll(query, "\n", ""))
	}
	return rx, tx, nil
}
