package http

import (
	"github.com/openziti-test-kitchen/zrok/util"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/config"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

func Run(cfg *Config) error {
	options := ziti.ListenOptions{
		ConnectTimeout: 5 * time.Minute,
		MaxConnections: 64,
	}
	zcfg, err := config.NewFromFile(cfg.IdentityPath)
	if err != nil {
		return errors.Wrap(err, "error loading config")
	}
	listener, err := ziti.NewContextWithConfig(zcfg).ListenWithOptions(cfg.Service, &options)
	if err != nil {
		return errors.Wrap(err, "error listening")
	}

	proxy, err := util.NewProxy(cfg.EndpointAddress)
	if err != nil {
		return err
	}
	if err := http.Serve(listener, util.NewProxyHandler(proxy)); err != nil {
		return err
	}

	return nil
}
