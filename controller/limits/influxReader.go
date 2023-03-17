package limits

import (
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/openziti/zrok/controller/metrics"
)

type influxReader struct {
	idb      influxdb2.Client
	queryApi api.QueryAPI
}

func newInfluxReader(cfg *metrics.InfluxConfig) *influxReader {
	idb := influxdb2.NewClient(cfg.Url, cfg.Token)
	queryApi := idb.QueryAPI(cfg.Org)
	return &influxReader{idb, queryApi}
}
