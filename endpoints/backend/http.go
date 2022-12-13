package backend

import (
	"context"
	"github.com/openziti-test-kitchen/zrok/util"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/config"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net"
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

type httpBind struct {
	cfg      *Config
	requests func() int32
	listener edge.Listener
	handler  http.Handler
}

func NewHTTP(cfg *Config) (*httpBind, error) {
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

	proxy, err := newReverseProxy(cfg.EndpointAddress)
	if err != nil {
		return nil, err
	}

	handler := util.NewProxyHandler(proxy)
	return &httpBind{
		cfg:      cfg,
		requests: handler.Requests,
		listener: listener,
		handler:  handler,
	}, nil
}

func (self *httpBind) Run() error {
	if err := http.Serve(self.listener, self.handler); err != nil {
		return err
	}
	return nil
}

func (self *httpBind) Requests() func() int32 {
	return self.requests
}

func newReverseProxy(target string) (*httputil.ReverseProxy, error) {
	targetURL, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	tpt := http.DefaultTransport.(*http.Transport).Clone()
	tpt.DialContext = metricsDial

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.Transport = tpt
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

func metricsDial(_ context.Context, network string, addr string) (net.Conn, error) {
	conn, err := net.Dial(network, addr)
	if err != nil {
		return conn, err
	}
	return newMetricsConn("backend", conn), nil
}
