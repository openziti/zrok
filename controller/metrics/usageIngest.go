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
					logrus.Errorf("unable to assert 'interval_start_utc': %v", event)
				}
			} else {
				logrus.Errorf("missing 'interval_start_utc': %v", event)
			}
			if v, found := eventMap["tags"]; found {
				if tags, ok := v.(map[string]interface{}); ok {
					if v, found := tags["serviceId"]; found {
						if vStr, ok := v.(string); ok {
							u.ZitiServiceId = vStr
						} else {
							logrus.Errorf("unable to assert 'tags/serviceId': %v", event)
						}
					} else {
						logrus.Errorf("missing 'tags/serviceId': %v", event)
					}
				} else {
					logrus.Errorf("unable to assert 'tags': %v", event)
				}
			} else {
				logrus.Errorf("missing 'tags': %v", event)
			}
			if v, found := eventMap["usage"]; found {
				if usage, ok := v.(map[string]interface{}); ok {
					if v, found := usage["ingress.tx"]; found {
						if vFloat64, ok := v.(float64); ok {
							u.FrontendTx = int64(vFloat64)
						} else {
							logrus.Errorf("unable to assert 'usage/ingress.tx': %v", event)
						}
					}
					if v, found := usage["ingress.rx"]; found {
						if vFloat64, ok := v.(float64); ok {
							u.FrontendRx = int64(vFloat64)
						} else {
							logrus.Errorf("unable to assert 'usage/ingress.rx': %v", event)
						}
					}
					if v, found := usage["egress.tx"]; found {
						if vFloat64, ok := v.(float64); ok {
							u.BackendTx = int64(vFloat64)
						} else {
							logrus.Errorf("unable to assert 'usage/egress.tx': %v", event)
						}
					}
					if v, found := usage["egress.rx"]; found {
						if vFloat64, ok := v.(float64); ok {
							u.BackendRx = int64(vFloat64)
						} else {
							logrus.Errorf("unable to assert 'usage/egress.rx': %v", event)
						}
					}
				} else {
					logrus.Errorf("unable to assert 'usage' (%v) %v", reflect.TypeOf(v), event)
				}
			} else {
				logrus.Warnf("missing 'usage': %v", event)
			}
			if v, found := eventMap["circuit_id"]; found {
				if vStr, ok := v.(string); ok {
					u.ZitiCircuitId = vStr
				} else {
					logrus.Errorf("unable to assert 'circuit_id': %v", event)
				}
			} else {
				logrus.Warnf("missing 'circuit_id': %v", event)
			}
		} else {
			logrus.Errorf("not 'fabric.usage': %v", event)
		}
		return u, nil
	} else {
		return nil, errors.Wrap(err, "error unmarshaling")
	}
}
