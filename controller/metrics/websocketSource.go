package metrics

import "github.com/michaelquigley/cf"

type WebsocketSourceConfig struct {
	WebsocketEndpoint string
}

func loadWebsocketSourceConfig(v interface{}, opts *cf.Options) (interface{}, error) {
	return nil, nil
}

type websocketSource struct{}

func (s *websocketSource) Start(events chan map[string]interface{}) (chan struct{}, error) {
	return nil, nil
}

func (s *websocketSource) Stop() {
}
