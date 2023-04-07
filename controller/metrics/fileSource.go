package metrics

import (
	"encoding/binary"
	"github.com/michaelquigley/cf"
	"github.com/nxadm/tail"
	"github.com/openziti/zrok/controller/env"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"os"
)

func init() {
	env.GetCfOptions().AddFlexibleSetter("fileSource", loadFileSourceConfig)
}

type FileSourceConfig struct {
	Path        string
	PointerPath string
}

func loadFileSourceConfig(v interface{}, _ *cf.Options) (interface{}, error) {
	if submap, ok := v.(map[string]interface{}); ok {
		cfg := &FileSourceConfig{}
		if err := cf.Bind(cfg, submap, cf.DefaultOptions()); err != nil {
			return nil, err
		}
		return &fileSource{cfg: cfg}, nil
	}
	return nil, errors.New("invalid config structure for 'fileSource'")
}

type fileSource struct {
	cfg  *FileSourceConfig
	ptrF *os.File
	t    *tail.Tail
}

func (s *fileSource) Start(events chan ZitiEventJson) (join chan struct{}, err error) {
	f, err := os.Open(s.cfg.Path)
	if err != nil {
		return nil, errors.Wrapf(err, "error opening '%v'", s.cfg.Path)
	}
	_ = f.Close()

	s.ptrF, err = os.OpenFile(s.pointerPath(), os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil, errors.Wrapf(err, "error opening pointer '%v'", s.pointerPath())
	}

	ptr, err := s.readPtr()
	if err != nil {
		logrus.Errorf("error reading pointer: %v", err)
	}
	logrus.Infof("retrieved stored position pointer at '%d'", ptr)

	join = make(chan struct{})
	go func() {
		s.tail(ptr, events)
		close(join)
	}()

	return join, nil
}

func (s *fileSource) Stop() {
	if err := s.t.Stop(); err != nil {
		logrus.Error(err)
	}
}

func (s *fileSource) tail(ptr int64, events chan ZitiEventJson) {
	logrus.Info("started")
	defer logrus.Info("stopped")

	var err error
	s.t, err = tail.TailFile(s.cfg.Path, tail.Config{
		ReOpen:   true,
		Follow:   true,
		Location: &tail.SeekInfo{Offset: ptr},
	})
	if err != nil {
		logrus.Errorf("error starting tail: %v", err)
		return
	}

	for event := range s.t.Lines {
		events <- ZitiEventJson(event.Text)

		if err := s.writePtr(event.SeekInfo.Offset); err != nil {
			logrus.Error(err)
		}
	}
}

func (s *fileSource) pointerPath() string {
	if s.cfg.PointerPath == "" {
		return s.cfg.Path + ".ptr"
	} else {
		return s.cfg.PointerPath
	}
}

func (s *fileSource) readPtr() (int64, error) {
	ptr := int64(0)
	buf := make([]byte, 8)
	if n, err := s.ptrF.Seek(0, 0); err == nil && n == 0 {
		if n, err := s.ptrF.Read(buf); err == nil && n == 8 {
			ptr = int64(binary.LittleEndian.Uint64(buf))
			return ptr, nil
		} else {
			return 0, errors.Wrapf(err, "error reading pointer (%d): %v", n, err)
		}
	} else {
		return 0, errors.Wrapf(err, "error seeking pointer (%d): %v", n, err)
	}
}

func (s *fileSource) writePtr(ptr int64) error {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(ptr))
	if n, err := s.ptrF.Seek(0, 0); err == nil && n == 0 {
		if n, err := s.ptrF.Write(buf); err != nil || n != 8 {
			return errors.Wrapf(err, "error writing pointer (%d): %v", n, err)
		}
	} else {
		return errors.Wrapf(err, "error seeking pointer (%d): %v", n, err)
	}
	return nil
}
