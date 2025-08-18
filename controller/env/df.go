package env

import (
	"github.com/michaelquigley/df"
	"github.com/openziti/zrok/controller/metrics"
)

var dfOpts *df.Options

func GetDfOptions() *df.Options {
	if dfOpts == nil {
		dfOpts = &df.Options{
			DynamicBinders: map[string]func(map[string]any) (df.Dynamic, error){
				metrics.AmqpSinkType:        metrics.LoadAmqpSink,
				metrics.AmqpSourceType:      metrics.LoadAmqpSource,
				metrics.FileSourceType:      metrics.LoadFileSource,
				metrics.WebsocketSourceType: metrics.LoadWebsocketSource,
			},
		}
	}
	return dfOpts
}
