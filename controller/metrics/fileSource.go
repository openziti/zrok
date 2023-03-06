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
		return &fileSource{cfg: cfg}, nil
	}
	return nil, errors.New("invalid config structure for 'file' source")
}

type fileSource struct {
	cfg *FileSourceConfig
	t   *tail.Tail
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
	if err := s.t.Stop(); err != nil {
		logrus.Error(err)
	}
}

func (s *fileSource) tail(events chan map[string]interface{}) {
	logrus.Infof("started")
	defer logrus.Infof("stopped")

	var err error
	s.t, err = tail.TailFile(s.cfg.Path, tail.Config{
		ReOpen: true,
		Follow: true,
	})
	if err != nil {
		logrus.Error(err)
		return
	}

	for line := range s.t.Lines {
		event := make(map[string]interface{})
		if err := json.Unmarshal([]byte(line.Text), &event); err == nil {
			logrus.Infof("seekinfo: offset: %d", line.SeekInfo.Offset)
			events <- event
		} else {
			logrus.Errorf("error parsing line #%d: %v", line.Num, err)
		}
	}
}
