package metrics

import "github.com/michaelquigley/cf"

type FileSourceConfig struct {
	Path string
}

func loadFileSourceConfig(v interface{}, opts *cf.Options) (interface{}, error) {
	return nil, nil
}

type fileSource struct{}

func (s *fileSource) Start() (chan struct{}, error) {
	return nil, nil
}
