package controller

import "github.com/openziti-test-kitchen/zrok/controller/store"

type Config struct {
	Endpoint EndpointConfig
	Proxy    ProxyConfig
	Store    *store.Config
}

type EndpointConfig struct {
	Host string
	Port int
}

type ProxyConfig struct {
	UrlTemplate string
}
