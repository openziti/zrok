package proxy

import (
	"fmt"
	"github.com/openziti/zrok/endpoints"
	"net/http"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

var middlewareRequests chan *endpoints.Request

func init() {
	caddy.RegisterModule(ZrokRequestsMiddleware{})
	httpcaddyfile.RegisterHandlerDirective("zrok_requests", parseCaddyfile)
}

type ZrokRequestsMiddleware struct{}

// CaddyModule returns the Caddy module information.
func (ZrokRequestsMiddleware) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.zrok_requests",
		New: func() caddy.Module { return new(ZrokRequestsMiddleware) },
	}
}

// Provision implements caddy.Provisioner.
func (m *ZrokRequestsMiddleware) Provision(ctx caddy.Context) error {
	return nil
}

// Validate implements caddy.Validator.
func (m ZrokRequestsMiddleware) Validate() error {
	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (m ZrokRequestsMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	if middlewareRequests != nil {
		middlewareRequests <- &endpoints.Request{
			Stamp:      time.Now(),
			RemoteAddr: fmt.Sprintf("%v", r.Header["X-Real-Ip"]),
			Method:     r.Method,
			Path:       r.URL.String(),
		}
	}
	return next.ServeHTTP(w, r)
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (m *ZrokRequestsMiddleware) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	return nil
}

// parseCaddyfile unmarshals tokens from h into a new ZrokRequestsMiddleware.
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var m ZrokRequestsMiddleware
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return m, err
}

// Interface guards
var (
	_ caddy.Provisioner           = (*ZrokRequestsMiddleware)(nil)
	_ caddy.Validator             = (*ZrokRequestsMiddleware)(nil)
	_ caddyhttp.MiddlewareHandler = (*ZrokRequestsMiddleware)(nil)
	_ caddyfile.Unmarshaler       = (*ZrokRequestsMiddleware)(nil)
)
