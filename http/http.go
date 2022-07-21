package http

import (
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/config"
	"github.com/pkg/errors"
	"net/http"
	"net/http/httputil"
	"net/url"
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
	listener, err := ziti.NewContextWithConfig(zcfg).ListenWithOptions("zrok", &options)
	if err != nil {
		return errors.Wrap(err, "error listening")
	}

	targetURL, err := url.Parse("http://localhost:3000")
	if err != nil {
		return errors.Wrap(err, "error parsing url")
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	if err := http.Serve(listener, &proxyHandler{proxy: proxy}); err != nil {
		return err
	}

	return nil
}

type proxyHandler struct {
	proxy *httputil.ReverseProxy
}

func (self *proxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	self.proxy.ServeHTTP(w, r)
}
