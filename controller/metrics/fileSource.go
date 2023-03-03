package metrics

import (
	"github.com/michaelquigley/cf"
	"github.com/pkg/errors"
	"os"
)

type FileSourceConfig struct {
	Path string
}

func loadFileSourceConfig(v interface{}, opts *cf.Options) (interface{}, error) {
	if submap, ok := v.(map[string]interface{}); ok {
		cfg := &FileSourceConfig{}
		if err := cf.Bind(cfg, submap, cf.DefaultOptions()); err != nil {
			return nil, err
		}
		return &fileSource{cfg}, nil
	}
	return nil, errors.New("invalid config structure for 'file' source")
}

type fileSource struct {
	cfg *FileSourceConfig
}

func (s *fileSource) Start() (chan struct{}, error) {
	f, err := os.Open(s.cfg.Path)
	if err != nil {
		return nil, errors.Wrapf(err, "error opening '%v'", s.cfg.Path)
	}
	ch := make(chan struct{})
	go func() {
		f.Close()
		close(ch)
	}()
	return ch, nil
}

func (s *fileSource) Stop() {
}
