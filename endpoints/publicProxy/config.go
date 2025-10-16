package publicProxy

import (
	"context"
	"net/http"
	"time"

	"github.com/michaelquigley/df/dd"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/endpoints"
	"github.com/openziti/zrok/endpoints/proxyUi"
	"github.com/pkg/errors"
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
	MaxCookieSize        int
	SessionLifetime      time.Duration
	IntermediateLifetime time.Duration
	SigningKey           string       `dd:"+secret"`
	EncryptionKey        string       `dd:"+secret"`
	Providers            []dd.Dynamic `dd:"+secret"`
}

func (c *OauthConfig) GetCookieName() string {
	return c.CookieName
}

func (c *OauthConfig) GetCookieDomain() string {
	return c.CookieDomain
}

func (c *OauthConfig) GetMaxCookieSize() int {
	return c.MaxCookieSize
}

func (c *OauthConfig) GetSessionLifetime() time.Duration {
	return c.SessionLifetime
}

func DefaultConfig() *Config {
	return &Config{
		Identity: "public",
		Address:  "0.0.0.0:8080",
		Oauth: &OauthConfig{
			MaxCookieSize: 3072,
		},
	}
}

func (c *Config) Load(path string) error {
	opts := &dd.Options{
		DynamicBinders: map[string]func(map[string]any) (dd.Dynamic, error){
			(&githubConfig{}).Type(): newGithubConfig,
			(&googleConfig{}).Type(): newGoogleConfig,
			(&oidcConfig{}).Type():   newOidcConfig,
		},
	}
	if err := dd.MergeFromYAML(c, path, opts); err != nil {
		return errors.Wrapf(err, "error loading frontend config '%v'", path)
	}
	if c.V != V {
		return errors.Errorf("invalid configuration version '%d'; expected '%d'", c.V, V)
	}
	return nil
}

func configureOauth(ctx context.Context, cfg *Config, tls bool) error {
	if cfg.Oauth == nil {
		dl.Info("no oauth configuration; skipping oauth handler startup")
		return nil
	}

	for _, v := range cfg.Oauth.Providers {
		if prvCfg, ok := v.(dd.Dynamic); ok {
			switch prvCfg.Type() {
			case "github":
				githubCfg, ok := prvCfg.(*githubConfig)
				if !ok {
					return errors.New("invalid github provider configuration")
				}
				if err := githubCfg.configure(cfg.Oauth, tls); err != nil {
					return err
				}

			case "google":
				googleCfg, ok := prvCfg.(*googleConfig)
				if !ok {
					return errors.New("invalid google provider configuration")
				}
				if err := googleCfg.configure(cfg.Oauth, tls); err != nil {
					return err
				}

			case "oidc":
				oidcCfg, ok := prvCfg.(*oidcConfig)
				if !ok {
					return errors.New("invalid oidc provider configuration")
				}
				if err := oidcCfg.configure(cfg.Oauth, tls); err != nil {
					return err
				}

			default:
				return errors.Errorf("invalid oauth provider type '%v'", prvCfg.Type())
			}
		} else {
			return errors.Errorf("invalid oauth provider configuration; missing 'type'")
		}
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData())
	})

	zhttp.StartServer(ctx, cfg.Oauth.BindAddress)

	return nil
}
