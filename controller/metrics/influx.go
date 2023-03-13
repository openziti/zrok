package metrics

import (
	"context"
	"fmt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/openziti/zrok/util"
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
	out := fmt.Sprintf("share: %v, circuit: %v", u.ShareToken, u.ZitiCircuitId)

	envId := fmt.Sprintf("%d", u.EnvironmentId)
	acctId := fmt.Sprintf("%d", u.AccountId)

	var pts []*write.Point
	circuitPt := influxdb2.NewPoint("circuits",
		map[string]string{"share": u.ShareToken, "envId": envId, "acctId": acctId},
		map[string]interface{}{"circuit": u.ZitiCircuitId},
		u.IntervalStart)
	pts = append(pts, circuitPt)

	if u.BackendTx > 0 || u.BackendRx > 0 {
		pt := influxdb2.NewPoint("xfer",
			map[string]string{"namespace": "backend", "share": u.ShareToken, "envId": envId, "acctId": acctId},
			map[string]interface{}{"bytesRead": u.BackendRx, "bytesWritten": u.BackendTx},
			u.IntervalStart)
		pts = append(pts, pt)
		out += fmt.Sprintf(" backend {rx: %v, tx: %v}", util.BytesToSize(u.BackendRx), util.BytesToSize(u.BackendTx))
	}
	if u.FrontendTx > 0 || u.FrontendRx > 0 {
		pt := influxdb2.NewPoint("xfer",
			map[string]string{"namespace": "frontend", "share": u.ShareToken, "envId": envId, "acctId": acctId},
			map[string]interface{}{"bytesRead": u.FrontendRx, "bytesWritten": u.FrontendTx},
			u.IntervalStart)
		pts = append(pts, pt)
		out += fmt.Sprintf(" frontend {rx: %v, tx: %v}", util.BytesToSize(u.FrontendRx), util.BytesToSize(u.FrontendTx))
	}

	if err := i.writeApi.WritePoint(context.Background(), pts...); err == nil {
		logrus.Info(out)
	} else {
		return err
	}

	return nil
}
