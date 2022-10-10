package frontend

import (
	"fmt"
	"github.com/openziti-test-kitchen/zrok/util"
	"time"
)

type metricsAgent struct {
	metrics map[string]sessionMetrics
	updates chan metricsUpdate
}

type sessionMetrics struct {
	bytesRead    int64
	bytesWritten int64
	lastUpdate   time.Time
}

type metricsUpdate struct {
	id           string
	bytesRead    int64
	bytesWritten int64
}

func newMetricsAgent() *metricsAgent {
	return &metricsAgent{
		metrics: make(map[string]sessionMetrics),
		updates: make(chan metricsUpdate, 10240),
	}
}

func (ma *metricsAgent) run() {
	for {
		select {
		case update := <-ma.updates:
			if sm, found := ma.metrics[update.id]; found {
				sm.bytesRead += update.bytesRead
				sm.bytesWritten += update.bytesWritten
				sm.lastUpdate = time.Now()
				ma.metrics[update.id] = sm
			} else {
				sm := sessionMetrics{
					bytesRead:    update.bytesRead,
					bytesWritten: update.bytesWritten,
					lastUpdate:   time.Now(),
				}
				ma.metrics[update.id] = sm
			}

		case <-time.After(5 * time.Second):
			now := time.Now()
			out := "metrics = {\n"
			for k, v := range ma.metrics {
				age := now.Sub(v.lastUpdate)
				out += fmt.Sprintf("\t[%v]: %s/%s (%s)\n", k, util.BytesToSize(v.bytesRead), util.BytesToSize(v.bytesWritten), age.String())
			}
			out += "}\n"
			//fmt.Println(out)
		}
	}
}
