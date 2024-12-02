package proxy

import (
	"crypto/tls"
	"fmt"
	"github.com/openziti/sdk-golang/ziti"
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

type BackendConfig struct {
	IdentityPath    string
	EndpointAddress string
	ShrToken        string
	Insecure        bool
	Requests        chan *endpoints.Request
}

type Backend struct {
	cfg      *BackendConfig
	listener edge.Listener
	handler  http.Handler
}

func NewBackend(cfg *BackendConfig) (*Backend, error) {
	options := ziti.ListenOptions{
		ConnectTimeout:               5 * time.Minute,
		WaitForNEstablishedListeners: 1,
	}
	zcfg, err := ziti.NewConfigFromFile(cfg.IdentityPath)
	if err != nil {
		return nil, errors.Wrap(err, "error loading config")
	}
	zctx, err := ziti.NewContext(zcfg)
	if err != nil {
		return nil, errors.Wrap(err, "error loading ziti context")
	}
	listener, err := zctx.ListenWithOptions(cfg.ShrToken, &options)
	if err != nil {
		return nil, errors.Wrap(err, "error listening")
	}

	proxy, err := newReverseProxy(cfg)
	if err != nil {
		return nil, err
	}

	handler := util.NewRequestsWrapper(proxy)
	return &Backend{
		cfg:      cfg,
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

func (b *Backend) Stop() error {
	return b.listener.Close()
}

func newReverseProxy(cfg *BackendConfig) (*httputil.ReverseProxy, error) {
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
		if cfg.Requests != nil {
			cfg.Requests <- &endpoints.Request{
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
		w.WriteHeader(http.StatusBadGateway)
	}

	return proxy, nil
}
