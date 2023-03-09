package metrics

import (
	"context"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/sirupsen/logrus"
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
	pt := influxdb2.NewPoint("xfer",
		map[string]string{"namespace": "backend", "share": u.ShareToken},
		map[string]interface{}{"bytesRead": u.BackendRx, "bytesWritten": u.BackendTx},
		u.IntervalStart)
	if err := i.writeApi.WritePoint(context.Background(), pt); err == nil {
		logrus.Infof("share: %v, circuit: %v, rx: %d, tx: %d", u.ShareToken, u.ZitiCircuitId, u.BackendRx, u.BackendTx)
	} else {
		return err
	}
	return nil
}
