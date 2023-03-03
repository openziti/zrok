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

	events := make(chan map[string]interface{}, 1024)
	srcJoin, err := src.Start(events)
	if err != nil {
		return errors.Wrap(err, "error starting source")
	}

	go func() {
		ingester := &UsageIngester{}
		for {
			select {
			case event := <-events:
				if err := ingester.Ingest(event); err != nil {
					logrus.Error(err)
				}
			}
		}
	}()

	<-srcJoin

	return nil
}
