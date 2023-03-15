package metrics

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"reflect"
	"time"
)

func Ingest(event ZitiEventJson) (*Usage, error) {
	eventMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(event), &eventMap); err == nil {
		u := &Usage{ProcessedStamp: time.Now()}
		if ns, found := eventMap["namespace"]; found && ns == "fabric.usage" {
			if v, found := eventMap["interval_start_utc"]; found {
				if vFloat64, ok := v.(float64); ok {
					u.IntervalStart = time.Unix(int64(vFloat64), 0)
				} else {
					logrus.Error("unable to assert 'interval_start_utc'")
				}
			} else {
				logrus.Error("missing 'interval_start_utc'")
			}
			if v, found := eventMap["tags"]; found {
				if tags, ok := v.(map[string]interface{}); ok {
					if v, found := tags["serviceId"]; found {
						if vStr, ok := v.(string); ok {
							u.ZitiServiceId = vStr
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
			if v, found := eventMap["usage"]; found {
				if usage, ok := v.(map[string]interface{}); ok {
					if v, found := usage["ingress.tx"]; found {
						if vFloat64, ok := v.(float64); ok {
							u.FrontendTx = int64(vFloat64)
						} else {
							logrus.Error("unable to assert 'usage/ingress.tx'")
						}
					} else {
						logrus.Warn("missing 'usage/ingress.tx'")
					}
					if v, found := usage["ingress.rx"]; found {
						if vFloat64, ok := v.(float64); ok {
							u.FrontendRx = int64(vFloat64)
						} else {
							logrus.Error("unable to assert 'usage/ingress.rx")
						}
					} else {
						logrus.Warn("missing 'usage/ingress.rx")
					}
					if v, found := usage["egress.tx"]; found {
						if vFloat64, ok := v.(float64); ok {
							u.BackendTx = int64(vFloat64)
						} else {
							logrus.Error("unable to assert 'usage/egress.tx'")
						}
					} else {
						logrus.Warn("missing 'usage/egress.tx'")
					}
					if v, found := usage["egress.rx"]; found {
						if vFloat64, ok := v.(float64); ok {
							u.BackendRx = int64(vFloat64)
						} else {
							logrus.Error("unable to assert 'usage/egress.rx'")
						}
					} else {
						logrus.Warn("missing 'usage/egress.rx'")
					}
				} else {
					logrus.Errorf("unable to assert 'usage' (%v) %v", reflect.TypeOf(v), event)
				}
			} else {
				logrus.Warnf("missing 'usage'")
			}
			if v, found := eventMap["circuit_id"]; found {
				if vStr, ok := v.(string); ok {
					u.ZitiCircuitId = vStr
				} else {
					logrus.Error("unable to assert 'circuit_id'")
				}
			} else {
				logrus.Warn("missing 'circuit_id'")
			}
		} else {
			logrus.Errorf("not 'fabric.usage'")
		}
		return u, nil
	} else {
		return nil, errors.Wrap(err, "error unmarshaling")
	}
}
