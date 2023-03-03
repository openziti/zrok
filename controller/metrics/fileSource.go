package metrics

import (
	"encoding/json"
	"github.com/michaelquigley/cf"
	"github.com/nxadm/tail"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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

func (s *fileSource) Start(events chan map[string]interface{}) (chan struct{}, error) {
	f, err := os.Open(s.cfg.Path)
	if err != nil {
		return nil, errors.Wrapf(err, "error opening '%v'", s.cfg.Path)
	}
	_ = f.Close()

	ch := make(chan struct{})
	go func() {
		s.tail(events)
		close(ch)
	}()

	return ch, nil
}

func (s *fileSource) Stop() {
}

func (s *fileSource) tail(events chan map[string]interface{}) {
	t, err := tail.TailFile(s.cfg.Path, tail.Config{Follow: true, ReOpen: true})
	if err != nil {
		logrus.Error(err)
		return
	}

	for line := range t.Lines {
		event := make(map[string]interface{})
		if err := json.Unmarshal([]byte(line.Text), &event); err == nil {
			events <- event
		} else {
			logrus.Errorf("error parsing line #%d: %v", line.Num, err)
		}
	}
}
