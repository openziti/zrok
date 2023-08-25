package proxy

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"io"
)

func init() {
	caddy.RegisterModule(CaddyLogWriter{})
}

func SetCaddyLoggingWriter(w io.WriteCloser) {
	loggingRequests = w
}

var loggingRequests io.WriteCloser

type CaddyLogWriter struct{}

func (CaddyLogWriter) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "caddy.logging.writers.zrok_tui",
		New: func() caddy.Module { return new(CaddyLogWriter) },
	}
}

func (w *CaddyLogWriter) Provision(_ caddy.Context) error {
	return nil
}

func (CaddyLogWriter) String() string {
	return ""
}

func (CaddyLogWriter) WriterKey() string {
	return "zrok_tui"
}

func (CaddyLogWriter) OpenWriter() (io.WriteCloser, error) {
	return loggingRequests, nil
}

func (*CaddyLogWriter) UnmarshalCaddyfile(_ *caddyfile.Dispenser) error {
	return nil
}

var (
	_ caddy.Provisioner     = (*CaddyLogWriter)(nil)
	_ caddy.WriterOpener    = (*CaddyLogWriter)(nil)
	_ caddyfile.Unmarshaler = (*CaddyLogWriter)(nil)
)
