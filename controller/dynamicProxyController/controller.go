package dynamicProxyController

import (
	"context"

	"github.com/michaelquigley/df"
	"github.com/openziti/zrok/sdk/golang/pubsub"
	"github.com/sirupsen/logrus"
)

type Operation string

const (
	OperationBind   Operation = "bind"
	OperationUnbind Operation = "unbind"
)

const MappingUpdate = "mapping-update"

type Mapping struct {
	Operation Operation
	Name      string
	Version   int64
}

type Controller struct {
	publisher pubsub.Publisher
}

func NewController(cfg *Config) (*Controller, error) {
	publisher, err := pubsub.NewPublisher(cfg.Publisher)
	if err != nil {
		return nil, err
	}
	return &Controller{publisher: publisher}, nil
}

func (c *Controller) SendMappingUpdate(frontendToken string, m Mapping) error {
	data, err := df.Unbind(m)
	if err != nil {
		return err
	}
	msg := pubsub.NewMessage(MappingUpdate, frontendToken, data)
	if err := c.publisher.Publish(context.Background(), msg); err != nil {
		return err
	}
	logrus.Infof("sent '%v' -> '%v'", data, frontendToken)
	return nil
}
