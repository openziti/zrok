package controller

import "github.com/openziti-test-kitchen/zrok/controller/store"

type Config struct {
	Host  string
	Port  int
	Store *store.Config
}
