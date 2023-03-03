package metrics

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func Run(cfg *Config) error {
	logrus.Info("starting")
	defer logrus.Warn("stopping")

	if cfg.Source == nil {
		return errors.New("no 'source' configured; exiting")
	}

	src, ok := cfg.Source.(Source)
	if !ok {
		return errors.New("invalid 'source'; exiting")
	}

	srcJoin, err := src.Start()
	if err != nil {
		return errors.Wrap(err, "error starting source")
	}

	<-srcJoin

	return nil
}
