package metrics

import (
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

type influxDb struct {
	idb      influxdb2.Client
	writeApi api.WriteAPIBlocking
}

func openInfluxDb(cfg *InfluxConfig) *influxDb {
	idb := influxdb2.NewClient(cfg.Url, cfg.Token)
	wapi := idb.WriteAPIBlocking(cfg.Org, cfg.Bucket)
	return &influxDb{idb, wapi}
}

func (i *influxDb) Write(u *Usage) error {
	return nil
}
