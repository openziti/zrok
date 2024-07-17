package proxy

import (
	_ "embed"
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp/fileserver"
	_ "github.com/greenpau/caddy-security"
	"github.com/openziti/zrok/endpoints"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
	"text/template"
)

//go:embed browse.html
var browseHtml string

func init() {
	fileserver.BrowseTemplate = browseHtml
}

type CaddyfileBackendConfig struct {
	CaddyfilePath string
	Shr           *sdk.Share
	Requests      chan *endpoints.Request
}

type CaddyfileBackend struct {
	cfg []byte
}

func NewCaddyfileBackend(cfg *CaddyfileBackendConfig) (*CaddyfileBackend, error) {
	cdyf, err := preprocessCaddyfile(cfg.CaddyfilePath, cfg.Shr)
	if err != nil {
		return nil, err
	}
	var adapter caddyfile.Adapter
	adapter.ServerType = httpcaddyfile.ServerType{}
	caddyCfg, warnings, err := adapter.Adapt([]byte(cdyf), map[string]interface{}{"filename": cfg.CaddyfilePath})
	if err != nil {
		return nil, err
	}
	for _, warning := range warnings {
		logrus.Warnf("%v [%d] (%v): %v", cfg.CaddyfilePath, warning.Line, warning.Directive, warning.Message)
	}
	return &CaddyfileBackend{cfg: caddyCfg}, nil
}

func (b *CaddyfileBackend) Run() error {
	if err := caddy.Load(b.cfg, true); err != nil {
		return err
	}
	return nil
}

func preprocessCaddyfile(inF string, shr *sdk.Share) (string, error) {
	input, err := os.ReadFile(inF)
	if err != nil {
		return "", err
	}
	tmpl, err := template.New(inF).Parse(string(input))
	if err != nil {
		return "", err
	}
	output := new(strings.Builder)
	if err := tmpl.Execute(output, &caddyfileData{ZrokBindAddress: fmt.Sprintf("zrok/%s", shr.Token)}); err != nil {
		return "", err
	}
	return output.String(), nil
}

type caddyfileData struct {
	ZrokBindAddress string
}
