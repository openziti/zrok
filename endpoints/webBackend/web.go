package webBackend

import (
	"fmt"
	"github.com/openziti-test-kitchen/zrok/endpoints"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/config"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

type Config struct {
	IdentityPath string
	WebRoot      string
	ShrToken     string
	RequestsChan chan *endpoints.Request
}

type backend struct {
	cfg      *Config
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

	be := &backend{
		cfg:      cfg,
		listener: listener,
	}
	if cfg.RequestsChan != nil {
		be.handler = &requestGrabber{requests: cfg.RequestsChan, handler: http.FileServer(http.Dir(cfg.WebRoot))}
	} else {
		be.handler = http.FileServer(http.Dir(cfg.WebRoot))
	}
	return be, nil
}

func (self *backend) Run() error {
	if err := http.Serve(self.listener, self.handler); err != nil {
		return err
	}
	return nil
}

func (self *backend) Requests() func() int32 {
	return func() int32 { return 0 }
}

type requestGrabber struct {
	requests chan *endpoints.Request
	handler  http.Handler
}

func (rl *requestGrabber) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if rl.requests != nil {
		rl.requests <- &endpoints.Request{
			Stamp:      time.Now(),
			RemoteAddr: fmt.Sprintf("%v", req.Header["X-Real-Ip"]),
			Method:     req.Method,
			Path:       req.URL.String(),
		}
	}
	rl.handler.ServeHTTP(resp, req)
}
