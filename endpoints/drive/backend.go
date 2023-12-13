package drive

import (
	"fmt"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/openziti/zrok/endpoints"
	"github.com/pkg/errors"
	"golang.org/x/net/webdav"
	"net/http"
	"time"
)

type BackendConfig struct {
	IdentityPath string
	DriveRoot    string
	ShrToken     string
	Requests     chan *endpoints.Request
}

type Backend struct {
	cfg      *BackendConfig
	listener edge.Listener
	handler  http.Handler
}

func NewBackend(cfg *BackendConfig) (*Backend, error) {
	options := ziti.ListenOptions{
		ConnectTimeout:               5 * time.Minute,
		MaxConnections:               64,
		WaitForNEstablishedListeners: 1,
	}
	zcfg, err := ziti.NewConfigFromFile(cfg.IdentityPath)
	if err != nil {
		return nil, errors.Wrap(err, "error loading ziti identity")
	}
	zctx, err := ziti.NewContext(zcfg)
	if err != nil {
		return nil, errors.Wrap(err, "error loading ziti context")
	}
	listener, err := zctx.ListenWithOptions(cfg.ShrToken, &options)
	if err != nil {
		return nil, err
	}

	handler := &webdav.Handler{
		FileSystem: webdav.Dir(cfg.DriveRoot),
		LockSystem: webdav.NewMemLS(),
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
