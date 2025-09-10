package config

import "github.com/openziti/zrok/endpoints/dynamicProxy/store"

const V = 1

type Config struct {
	V              int                   `df:"required"`
	Store          *store.Config         `df:"required"`
	AmqpSubscriber *AmqpSubscriberConfig `df:"required"`
}

type AmqpSubscriberConfig struct {
	Url           string `df:"required"`
	ExchangeName  string `df:"required"`
	FrontendToken string `df:"required"`
}
