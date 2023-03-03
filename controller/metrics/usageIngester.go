package metrics

import (
	"github.com/openziti/zrok/util"
	"github.com/sirupsen/logrus"
	"reflect"
)

type UsageIngester struct{}

func (i *UsageIngester) Ingest(event map[string]interface{}) error {
	if ns, found := event["namespace"]; found && ns == "fabric.usage" {
		start := float64(0)
		if v, found := event["interval_start_utc"]; found {
			if vFloat64, ok := v.(float64); ok {
				start = vFloat64
			} else {
				logrus.Error("unable to assert 'interval_start_utc'")
			}
		} else {
			logrus.Error("missing 'interval_start_utc'")
		}
		clientId := ""
		serviceId := ""
		if v, found := event["tags"]; found {
			if tags, ok := v.(map[string]interface{}); ok {
				if v, found := tags["clientId"]; found {
					if vStr, ok := v.(string); ok {
						clientId = vStr
					} else {
						logrus.Error("unable to assert 'tags/clientId'")
					}
				} else {
					logrus.Errorf("missing 'tags/clientId'")
				}
				if v, found := tags["serviceId"]; found {
					if vStr, ok := v.(string); ok {
						serviceId = vStr
					} else {
						logrus.Error("unable to assert 'tags/serviceId'")
					}
				} else {
					logrus.Error("missing 'tags/serviceId'")
				}
			} else {
				logrus.Errorf("unable to assert 'tags'")
			}
		} else {
			logrus.Errorf("missing 'tags'")
		}
		tx := int64(0)
		rx := int64(0)
		if v, found := event["usage"]; found {
			if usage, ok := v.(map[string]interface{}); ok {
				if v, found := usage["egress.tx"]; found {
					if vFloat64, ok := v.(float64); ok {
						tx = int64(vFloat64)
					} else {
						logrus.Error("unable to assert 'usage/egress.tx'")
					}
				} else {
					logrus.Error("missing 'usage/egress.tx'")
				}
				if v, found := usage["egress.rx"]; found {
					if vFloat64, ok := v.(float64); ok {
						rx = int64(vFloat64)
					} else {
						logrus.Error("unable to assert 'usage/egress.rx'")
					}
				} else {
					logrus.Error("missing 'usage/egress.rx'")
				}
			} else {
				logrus.Errorf("unable to assert 'usage' (%v) %v", reflect.TypeOf(v), event)
			}
		} else {
			logrus.Error("missing 'usage'")
		}
		circuitId := ""
		if v, found := event["circuit_id"]; found {
			if vStr, ok := v.(string); ok {
				circuitId = vStr
			} else {
				logrus.Error("unable to assert 'circuit_id'")
			}
		} else {
			logrus.Error("missing 'circuit_id'")
		}

		logrus.Infof("usage: start '%d', serviceId '%v', clientId '%v', circuitId '%v' [rx: %v, tx: %v]", int64(start), serviceId, clientId, circuitId, util.BytesToSize(rx), util.BytesToSize(tx))

	} else {
		logrus.Errorf("not 'fabric.usage'")
	}
	return nil
}
