package drive

import (
	"fmt"
	"net/http"
	"time"

	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/openziti/zrok/drives/davServer"
	"github.com/openziti/zrok/endpoints"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type BackendConfig struct {
	IdentityPath string
	DriveRoot    string
	ShrToken     string
	Requests     chan *endpoints.Request
	SuperNetwork bool
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
		return nil, errors.Wrap(err, "error loading ziti identity")
	}
	if cfg.SuperNetwork {
		zcfg.MaxDefaultConnections = 2
		zcfg.MaxControlConnections = 1
		logrus.Warnf("super networking enabled")
	}
	zctx, err := ziti.NewContext(zcfg)
	if err != nil {
		return nil, errors.Wrap(err, "error loading ziti context")
	}
	listener, err := zctx.ListenWithOptions(cfg.ShrToken, &options)
	if err != nil {
		return nil, err
	}

	handler := &davServer.Handler{
		FileSystem: davServer.Dir(cfg.DriveRoot),
		LockSystem: davServer.NewMemLS(),
		Logger: func(r *http.Request, err error) {
			if cfg.Requests != nil {
				cfg.Requests <- &endpoints.Request{
					Stamp:      time.Now(),
					RemoteAddr: fmt.Sprintf("%v", r.Header["X-Real-Ip"]),
					Method:     r.Method,
					Path:       r.URL.String(),
				}
			}
		},
	}

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
