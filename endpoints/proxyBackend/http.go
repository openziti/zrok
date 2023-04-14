package proxyBackend

import (
	"crypto/tls"
	"fmt"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/config"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/openziti/zrok/endpoints"
	"github.com/openziti/zrok/util"
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
	ShrToken        string
	Insecure        bool
	RequestsChan    chan *endpoints.Request
}

type Backend struct {
	cfg      *Config
	requests func() int32
	listener edge.Listener
	handler  http.Handler
}

func New(cfg *Config) (*Backend, error) {
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

	proxy, err := newReverseProxy(cfg)
	if err != nil {
		return nil, err
	}

	handler := util.NewProxyHandler(proxy)
	return &Backend{
		cfg:      cfg,
		requests: handler.Requests,
		listener: listener,
		handler:  handler,
	}, nil
}

func (b *Backend) Run() error {
	if err := http.Serve(b.listener, b.handler); err != nil {
		return err
	}
	return nil
}

func (b *Backend) Requests() func() int32 {
	return b.requests
}

func newReverseProxy(cfg *Config) (*httputil.ReverseProxy, error) {
	targetURL, err := url.Parse(cfg.EndpointAddress)
	if err != nil {
		return nil, err
	}

	tpt := http.DefaultTransport.(*http.Transport).Clone()
	if cfg.Insecure {
		tpt.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.Transport = tpt
	director := proxy.Director
	proxy.Director = func(req *http.Request) {
		if cfg.RequestsChan != nil {
			cfg.RequestsChan <- &endpoints.Request{
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
