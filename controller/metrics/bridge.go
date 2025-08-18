package metrics

import (
	"github.com/michaelquigley/df"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type BridgeConfig struct {
	Source df.Dynamic
	Sink   df.Dynamic
}

type Bridge struct {
	src     ZitiEventJsonSource
	srcJoin chan struct{}
	snk     ZitiEventJsonSink
	events  chan ZitiEventMsg
	close   chan struct{}
	join    chan struct{}
}

func NewBridge(cfg *BridgeConfig) (*Bridge, error) {
	b := &Bridge{
		events: make(chan ZitiEventMsg),
		join:   make(chan struct{}),
		close:  make(chan struct{}),
	}
	if v, ok := cfg.Source.(ZitiEventJsonSource); ok {
		b.src = v
	} else {
		return nil, errors.New("invalid source type")
	}
	if v, ok := cfg.Sink.(ZitiEventJsonSink); ok {
		b.snk = v
	} else {
		return nil, errors.New("invalid sink type")
	}
	return b, nil
}

func (b *Bridge) Start() (join chan struct{}, err error) {
	if b.srcJoin, err = b.src.Start(b.events); err != nil {
		return nil, err
	}

	go func() {
		logrus.Info("started")
		defer logrus.Info("stopped")
		defer close(b.join)

	eventLoop:
		for {
			select {
			case eventJson := <-b.events:
				logrus.Info(eventJson)
				if err := b.snk.Handle(eventJson.Data()); err == nil {
					logrus.Infof("-> %v", eventJson.Data())
				} else {
					logrus.Error(err)
				}
				eventJson.Ack()

			case <-b.close:
				logrus.Info("received close signal")
				break eventLoop
			}
		}
	}()

	return b.join, nil
}

func (b *Bridge) Stop() {
	b.src.Stop()
	close(b.close)
	<-b.srcJoin
	<-b.join
}
