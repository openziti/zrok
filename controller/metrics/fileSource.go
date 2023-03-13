package metrics

import (
	"encoding/binary"
	"encoding/json"
	"github.com/michaelquigley/cf"
	"github.com/nxadm/tail"
	"github.com/openziti/zrok/controller/env"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"os"
)

func init() {
	env.GetCfOptions().AddFlexibleSetter("file", loadFileSourceConfig)
}

type FileSourceConfig struct {
	Path      string
	IndexPath string
}

func loadFileSourceConfig(v interface{}, _ *cf.Options) (interface{}, error) {
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

func (s *fileSource) Start(events chan map[string]interface{}) (join chan struct{}, err error) {
	f, err := os.Open(s.cfg.Path)
	if err != nil {
		return nil, errors.Wrapf(err, "error opening '%v'", s.cfg.Path)
	}
	_ = f.Close()

	idxF, err := os.OpenFile(s.indexPath(), os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil, errors.Wrapf(err, "error opening '%v'", s.indexPath())
	}

	pos := int64(0)
	posBuf := make([]byte, 8)
	if n, err := idxF.Read(posBuf); err == nil && n == 8 {
		pos = int64(binary.LittleEndian.Uint64(posBuf))
		logrus.Infof("recovered stored position: %d", pos)
	}

	join = make(chan struct{})
	go func() {
		s.tail(pos, events, idxF)
		close(join)
	}()

	return join, nil
}

func (s *fileSource) Stop() {
	if err := s.t.Stop(); err != nil {
		logrus.Error(err)
	}
}

func (s *fileSource) tail(pos int64, events chan map[string]interface{}, idxF *os.File) {
	logrus.Infof("started")
	defer logrus.Infof("stopped")

	posBuf := make([]byte, 8)

	var err error
	s.t, err = tail.TailFile(s.cfg.Path, tail.Config{
		ReOpen:   true,
		Follow:   true,
		Location: &tail.SeekInfo{Offset: pos},
	})
	if err != nil {
		logrus.Error(err)
		return
	}

	for line := range s.t.Lines {
		event := make(map[string]interface{})
		if err := json.Unmarshal([]byte(line.Text), &event); err == nil {
			binary.LittleEndian.PutUint64(posBuf, uint64(line.SeekInfo.Offset))
			if n, err := idxF.Seek(0, 0); err == nil && n == 0 {
				if n, err := idxF.Write(posBuf); err != nil || n != 8 {
					logrus.Errorf("error writing index (%d): %v", n, err)
				}
			}
			events <- event
		} else {
			logrus.Errorf("error parsing line #%d: %v", line.Num, err)
		}
	}
}

func (s *fileSource) indexPath() string {
	if s.cfg.IndexPath == "" {
		return s.cfg.Path + ".idx"
	} else {
		return s.cfg.IndexPath
	}
}
