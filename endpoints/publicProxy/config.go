package publicProxy

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/michaelquigley/cf"
	"github.com/openziti/zrok/endpoints"
	"github.com/openziti/zrok/endpoints/proxyUi"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	zhttp "github.com/zitadel/oidc/v2/pkg/http"
)

const V = 4

type Config struct {
	V            int
	Identity     string
	Address      string
	HostMatch    string
	TemplatePath string
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
	BindAddress          string
	EndpointUrl          string
	CookieName           string
	CookieDomain         string
	SessionLifetime      time.Duration
	IntermediateLifetime time.Duration
	MaxCookieSize        int
	SigningKey           string        `cf:"+secret"`
	EncryptionKey        string        `cf:"+secret"`
	Providers            []interface{} `cf:"+secret"`
}

type OauthProviderConfig struct {
	Name         string
	ClientId     string
	ClientSecret string `cf:"+secret"`
}

func DefaultConfig() *Config {
	return &Config{
		Identity: "public",
		Address:  "0.0.0.0:8080",
	}
}

func (c *Config) Load(path string) error {
	if err := cf.BindYaml(c, path, cfOptions()); err != nil {
		return errors.Wrapf(err, "error loading frontend config '%v'", path)
	}
	if c.V != V {
		return errors.Errorf("invalid configuration version '%d'; expected '%d'", c.V, V)
	}
	return nil
}

func cfOptions() *cf.Options {
	cfOpts := cf.DefaultOptions()
	cfOpts.AddFlexibleSetter("github", func(v interface{}, opt *cf.Options) (interface{}, error) {
		if vm, ok := v.(map[string]interface{}); ok {
			return vm, nil
		}
		return nil, fmt.Errorf("expected 'map[string]interface{}' got '%T'", v)
	})
	cfOpts.AddFlexibleSetter("google", func(v interface{}, opt *cf.Options) (interface{}, error) {
		if vm, ok := v.(map[string]interface{}); ok {
			return vm, nil
		}
		return nil, fmt.Errorf("expected 'map[string]interface{}' got '%T'", v)
	})
	cfOpts.AddFlexibleSetter("oidc", func(v interface{}, opt *cf.Options) (interface{}, error) {
		if vm, ok := v.(map[string]interface{}); ok {
			return vm, nil
		}
		return nil, fmt.Errorf("expected 'map[string]interface{}' got '%T'", v)
	})
	return cfOpts
}

func configureOauth(ctx context.Context, cfg *Config, tls bool) error {
	if cfg.Oauth == nil {
		logrus.Info("no oauth configuration; skipping oauth handler startup")
		return nil
	}

	for _, v := range cfg.Oauth.Providers {
		if mv, ok := v.(map[string]interface{}); ok {
			if t, found := mv["type"]; found {
				switch t {
				case "github":
					cfger, err := newGithubConfigurer(cfg.Oauth, tls, mv)
					if err != nil {
						return err
					}
					if err := cfger.configure(); err != nil {
						return err
					}

				case "google":
					cfger, err := newGoogleConfigurer(cfg.Oauth, tls, mv)
					if err != nil {
						return err
					}
					if err := cfger.configure(); err != nil {
						return err
					}

				case "oidc":
					cfger, err := newOidcConfigurer(cfg.Oauth, tls, mv)
					if err != nil {
						return err
					}
					if err := cfger.configure(); err != nil {
						return err
					}

				default:
					return errors.Errorf("invalid oauth provider type '%v'", t)
				}
			} else {
				return errors.Errorf("invalid oauth provider configuration; missing 'type'")
			}
		} else {
			return errors.Errorf("invalid oauth provider configuration data type")
		}
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData())
	})

	zhttp.StartServer(ctx, cfg.Oauth.BindAddress)

	return nil
}

func (c *OauthConfig) GetCookieName() string             { return c.CookieName }
func (c *OauthConfig) GetCookieDomain() string           { return c.CookieDomain }
func (c *OauthConfig) GetMaxCookieSize() int             { return c.MaxCookieSize }
func (c *OauthConfig) GetSessionLifetime() time.Duration { return c.SessionLifetime }
