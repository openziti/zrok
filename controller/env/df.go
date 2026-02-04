package env

import (
	"github.com/michaelquigley/df/dd"
	"github.com/openziti/zrok/v2/controller/metrics"
)

var ddOpts *dd.Options

func GetDdOptions() *dd.Options {
	if ddOpts == nil {
		ddOpts = &dd.Options{
			DynamicBinders: map[string]func(map[string]any) (dd.Dynamic, error){
				metrics.AmqpSinkType:        metrics.LoadAmqpSink,
				metrics.AmqpSourceType:      metrics.LoadAmqpSource,
				metrics.FileSourceType:      metrics.LoadFileSource,
				metrics.WebsocketSourceType: metrics.LoadWebsocketSource,
			},
		}
	}
	return ddOpts
}
