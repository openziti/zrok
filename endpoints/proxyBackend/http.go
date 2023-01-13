package proxyBackend

import (
	"context"
	"fmt"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/config"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/openziti/zrok/endpoints"
	"github.com/openziti/zrok/util"
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
	ShrToken        string
	RequestsChan    chan *endpoints.Request
}

type backend struct {
	cfg      *Config
	requests func() int32
	listener edge.Listener
	handler  http.Handler
}

func NewBackend(cfg *Config) (*backend, error) {
	options := ziti.ListenOptions{
		ConnectTimeout: 5 * time.Minute,
		MaxConnections: 64,
	}
	zcfg, err := config.NewFromFile(cfg.IdentityPath)
	if err != nil {
		return nil, errors.Wrap(err, "error loading config")
	}
	listener, err := ziti.NewContextWithConfig(zcfg).ListenWithOptions(cfg.ShrToken, &options)
	if err != nil {
		return nil, errors.Wrap(err, "error listening")
	}

	proxy, err := newReverseProxy(cfg.EndpointAddress, cfg.RequestsChan)
	if err != nil {
		return nil, err
	}

	handler := util.NewProxyHandler(proxy)
	return &backend{
		cfg:      cfg,
		requests: handler.Requests,
		listener: listener,
		handler:  handler,
	}, nil
}

func (self *backend) Run() error {
	if err := http.Serve(self.listener, self.handler); err != nil {
		return err
	}
	return nil
}

func (self *backend) Requests() func() int32 {
	return self.requests
}

func newReverseProxy(target string, requests chan *endpoints.Request) (*httputil.ReverseProxy, error) {
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
		if requests != nil {
			requests <- &endpoints.Request{
				Stamp:      time.Now(),
				RemoteAddr: fmt.Sprintf("%v", req.Header["X-Real-Ip"]),
				Method:     req.Method,
				Path:       req.URL.String(),
			}
		}
		director(req)
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
