package proxy

import (
	"encoding/json"
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp/fileserver"
	"github.com/openziti/zrok/endpoints"
	"go.uber.org/zap"
	"time"
)

type CaddyWebBackendConfig struct {
	IdentityPath string
	WebRoot      string
	ShrToken     string
	Requests     chan *endpoints.Request
}

type CaddyWebBackend struct {
	cfg      *CaddyWebBackendConfig
	caddyCfg *caddy.Config
}

func NewCaddyWebBackend(cfg *CaddyWebBackendConfig) (*CaddyWebBackend, error) {
	handler := fileserver.FileServer{Root: cfg.WebRoot}
	handler.Browse = new(fileserver.Browse)

	var handlers []json.RawMessage
	middlewareRequests = cfg.Requests
	handlers = append(handlers, caddyconfig.JSONModuleObject(&ZrokRequestsMiddleware{}, "handler", "zrok_requests", nil))
	handlers = append(handlers, caddyconfig.JSONModuleObject(handler, "handler", "file_server", nil))

	route := caddyhttp.Route{HandlersRaw: handlers}

	server := &caddyhttp.Server{
		ReadHeaderTimeout: caddy.Duration(10 * time.Second),
		IdleTimeout:       caddy.Duration(30 * time.Second),
		MaxHeaderBytes:    1024 * 10,
		Routes:            caddyhttp.RouteList{route},
	}
	server.Listen = []string{fmt.Sprintf("zrok/%s", cfg.ShrToken)}

	httpApp := caddyhttp.App{
		Servers: map[string]*caddyhttp.Server{"static": server},
	}

	var false bool
	caddyCfg := &caddy.Config{
		Admin: &caddy.AdminConfig{
			Disabled: true,
			Config: &caddy.ConfigSettings{
				Persist: &false,
			},
		},
		AppsRaw: caddy.ModuleMap{
			"http": caddyconfig.JSON(httpApp, nil),
		},
		Logging: &caddy.Logging{
			Logs: map[string]*caddy.CustomLog{
				"default": {
					BaseLog: caddy.BaseLog{
						Level: zap.ErrorLevel.CapitalString(),
					},
				},
			},
		},
	}
	if loggingRequests != nil {
		caddyLog := caddyCfg.Logging.Logs["default"]
		caddyLog.WriterRaw = caddyconfig.JSONModuleObject(&CaddyLogWriter{}, "output", "zrok_tui", nil)
		caddyCfg.Logging.Logs["default"] = caddyLog
	}

	return &CaddyWebBackend{cfg: cfg, caddyCfg: caddyCfg}, nil
}

func (c *CaddyWebBackend) Run() error {
	return caddy.Run(c.caddyCfg)
}

func (c *CaddyWebBackend) Stop() error {
	return caddy.Stop()
}

func (c *CaddyWebBackend) Requests() func() int32 {
	return func() int32 { return 0 }
}
