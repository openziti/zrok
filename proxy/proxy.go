package proxy

import (
	"github.com/openziti-test-kitchen/zrok/util"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/config"
	"github.com/pkg/errors"
	"net/http"
	"net/http/httputil"
	"net/url"
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

	targetURL, err := url.Parse("http://zrok")
	if err != nil {
		return errors.Wrap(err, "error parsing url")
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.Transport = zTransport
	return http.ListenAndServe(cfg.Address, &proxyHandler{proxy: proxy})
}

type proxyHandler struct {
	proxy *httputil.ReverseProxy
}

func (self *proxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	self.proxy.ServeHTTP(w, r)
}
