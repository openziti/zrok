package metrics

import (
	"context"
	"fmt"
	"hash/fnv"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/util"
)

type influxWriter struct {
	idb          influxdb2.Client
	writeApi     api.WriteAPIBlocking
	writeTimeout time.Duration
}

func newInfluxWriter(cfg *InfluxConfig) *influxWriter {
	idb := influxdb2.NewClientWithOptions(cfg.Url, cfg.Token, influxdb2.DefaultOptions().SetMaxRetries(0))
	writeApi := idb.WriteAPIBlocking(cfg.Org, cfg.Bucket)
	timeout := cfg.WriteTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &influxWriter{idb: idb, writeApi: writeApi, writeTimeout: timeout}
}

func (w *influxWriter) Handle(u *Usage) error {
	return w.HandleContext(context.Background(), u)
}

func (w *influxWriter) HandleContext(parent context.Context, u *Usage) error {
	ctx, cancel := context.WithTimeout(parent, w.writeTimeout)
	defer cancel()
	if u.ShareToken != "" {
		out := fmt.Sprintf("share: %v, circuit: %v", u.ShareToken, u.ZitiCircuitId)

		envId := fmt.Sprintf("%d", u.EnvironmentId)
		acctId := fmt.Sprintf("%d", u.AccountId)
		pointTime := usagePointTime(u)

		var pts []*write.Point
		circuitPt := influxdb2.NewPoint("circuits",
			map[string]string{"share": u.ShareToken, "envId": envId, "acctId": acctId},
			map[string]interface{}{"circuit": u.ZitiCircuitId},
			pointTime)
		pts = append(pts, circuitPt)

		if u.BackendTx > 0 || u.BackendRx > 0 {
			pt := influxdb2.NewPoint("xfer",
				map[string]string{"namespace": "backend", "share": u.ShareToken, "envId": envId, "acctId": acctId},
				map[string]interface{}{"rx": u.BackendRx, "tx": u.BackendTx},
				pointTime)
			pts = append(pts, pt)
			out += fmt.Sprintf(" backend {rx: %v, tx: %v}", util.BytesToSize(u.BackendRx), util.BytesToSize(u.BackendTx))
		}
		if u.FrontendTx > 0 || u.FrontendRx > 0 {
			pt := influxdb2.NewPoint("xfer",
				map[string]string{"namespace": "frontend", "share": u.ShareToken, "envId": envId, "acctId": acctId},
				map[string]interface{}{"rx": u.FrontendRx, "tx": u.FrontendTx},
				pointTime)
			pts = append(pts, pt)
			out += fmt.Sprintf(" frontend {rx: %v, tx: %v}", util.BytesToSize(u.FrontendRx), util.BytesToSize(u.FrontendTx))
		}

		if err := w.writeApi.WritePoint(ctx, pts...); err == nil {
			dl.Info(out)
		} else {
			return err
		}
	}

	return nil
}

// usagePointTime must stay identical in every consumer and across releases:
// redelivering an event must replace its point rather than count it twice.
func usagePointTime(u *Usage) time.Time {
	h := fnv.New64a()
	for _, part := range []string{u.ZitiCircuitId, u.ZitiRouterId, u.ZitiReportTime} {
		_, _ = h.Write([]byte(part))
		_, _ = h.Write([]byte{0})
	}
	return u.IntervalStart.Add(time.Duration(h.Sum64()%999999999 + 1))
}
