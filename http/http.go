package http

import (
	"github.com/openziti-test-kitchen/zrok/util"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/config"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type Config struct {
	IdentityPath    string
	EndpointAddress string
	Service         string
}

type httpProxy struct {
	Requests func() int32
	listener edge.Listener
	handler  http.Handler
}

func New(cfg *Config) (*httpProxy, error) {
	options := ziti.ListenOptions{
		ConnectTimeout: 5 * time.Minute,
		MaxConnections: 64,
	}
	zcfg, err := config.NewFromFile(cfg.IdentityPath)
	if err != nil {
		return nil, errors.Wrap(err, "error loading config")
	}
	listener, err := ziti.NewContextWithConfig(zcfg).ListenWithOptions(cfg.Service, &options)
	if err != nil {
		return nil, errors.Wrap(err, "error listening")
	}

	proxy, err := NewProxy(cfg.EndpointAddress)
	if err != nil {
		return nil, err
	}

	handler := util.NewProxyHandler(proxy)
	return &httpProxy{
		Requests: handler.Requests,
		listener: listener,
		handler:  handler,
	}, nil
}

func (p *httpProxy) Run() error {
	if err := http.Serve(p.listener, p.handler); err != nil {
		return err
	}
	return nil
}

func NewProxy(target string) (*httputil.ReverseProxy, error) {
	targetURL, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	director := proxy.Director
	proxy.Director = func(req *http.Request) {
		director(req)
		logrus.Debugf("-> %v", req.URL.String())
		req.Header.Set("X-Proxy", "zrok")
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logrus.Errorf("error proxying: %v", err)
	}

	return proxy, nil
}
