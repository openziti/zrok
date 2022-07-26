package proxy

import (
	"github.com/openziti-test-kitchen/zrok/util"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

func Run(cfg *Config) error {
	zCfg, err := config.NewFromFile(cfg.IdentityPath)
	if err != nil {
		return errors.Wrap(err, "error loading config")
	}
	zCtx := ziti.NewContextWithConfig(zCfg)
	zDialCtx := util.ZitiDialContext{Context: zCtx}
	zTransport := http.DefaultTransport.(*http.Transport).Clone()
	zTransport.DialContext = zDialCtx.Dial

	proxy, err := util.NewServiceProxy(&resolver{})
	if err != nil {
		return err
	}
	proxy.Transport = zTransport
	return http.ListenAndServe(cfg.Address, util.NewProxyHandler(proxy))
}

type resolver struct {
}

func (r *resolver) Service(host string) string {
	logrus.Infof("host = '%v'", host)
	tokens := strings.Split(host, ".")
	if len(tokens) > 0 {
		return tokens[0]
	}
	return "zrok"
}
