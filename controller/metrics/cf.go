package metrics

import "github.com/michaelquigley/cf"

func GetCfOptions() *cf.Options {
	opts := cf.DefaultOptions()
	opts.AddFlexibleSetter("file", loadFileSourceConfig)
	opts.AddFlexibleSetter("websocket", loadWebsocketSourceConfig)
	return opts
}
