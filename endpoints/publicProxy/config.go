package publicProxy

import (
	"context"
	"crypto/md5"

	"github.com/michaelquigley/cf"
	"github.com/openziti/zrok/endpoints"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	zhttp "github.com/zitadel/oidc/v2/pkg/http"
	"golang.org/x/oauth2"
)

const V = 3

type Config struct {
	V            int
	Identity     string
	Address      string
	HostMatch    string
	Interstitial *InterstitialConfig
	Oauth        *OauthConfig
	Tls          *endpoints.TlsConfig
}

type InterstitialConfig struct {
	Enabled           bool
	HtmlPath          string
	UserAgentPrefixes []string
}

type OauthConfig struct {
	BindAddress  string
	RedirectUrl  string
	CookieDomain string
	HashKey      string `cf:"+secret"`
	Providers    []*OauthProviderConfig
}

type OauthProviderConfig struct {
	Name          string
	ClientId      string
	ClientSecret  string `cf:"+secret"`
	Scopes        []string
	AuthURL       string
	TokenURL      string
	EmailEndpoint string
	EmailPath     string
	SupportsPKCE  bool
}

func (p *OauthProviderConfig) GetEndpoint() oauth2.Endpoint {
	return oauth2.Endpoint{
		AuthURL:  p.AuthURL,
		TokenURL: p.TokenURL,
	}
}

func DefaultConfig() *Config {
	return &Config{
		Identity: "public",
		Address:  "0.0.0.0:8080",
	}
}

func (c *Config) Load(path string) error {
	if err := cf.BindYaml(c, path, cf.DefaultOptions()); err != nil {
		return errors.Wrapf(err, "error loading frontend config '%v'", path)
	}
	if c.V != V {
		return errors.Errorf("invalid configuration version '%d'; expected '%d'", c.V, V)
	}
	return nil
}

func configureOauthHandlers(ctx context.Context, cfg *Config, tls bool) error {
	if cfg.Oauth == nil {
		logrus.Info("no oauth configuration; skipping oauth handler startup")
		return nil
	}

	hash := md5.New()
	if n, err := hash.Write([]byte(cfg.Oauth.HashKey)); err != nil {
		return err
	} else if n != len(cfg.Oauth.HashKey) {
		return errors.New("short hash")
	}
	key := hash.Sum(nil)

	for _, providerCfg := range cfg.Oauth.Providers {
		provider, err := configureOIDCProvider(cfg.Oauth, providerCfg, tls)
		if err != nil {
			logrus.Warnf("failed to configure provider %s: %v", providerCfg.Name, err)
			continue
		}
		provider.setupHandlers(cfg.Oauth, key, tls)
	}

	zhttp.StartServer(ctx, cfg.Oauth.BindAddress)
	return nil
}
