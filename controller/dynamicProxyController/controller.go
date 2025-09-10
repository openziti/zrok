package dynamicProxyController

import (
	"context"

	"github.com/openziti/zrok/dynamicProxyModel"
	"github.com/sirupsen/logrus"
)

type Controller struct {
	publisher *AmqpPublisher
}

func NewController(cfg *Config) (*Controller, error) {
	publisher, err := NewAmqpPublisher(cfg.AmqpPublisher)
	if err != nil {
		return nil, err
	}
	return &Controller{publisher: publisher}, nil
}

func (c *Controller) SendMappingUpdate(frontendToken string, m dynamicProxyModel.Mapping) error {
	if err := c.publisher.Publish(context.Background(), frontendToken, m); err != nil {
		return err
	}
	logrus.Infof("sent mapping update '%+v' -> '%s'", m, frontendToken)
	return nil
}
