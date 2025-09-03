package publicProxy

import (
	"context"
	"net/http"
	"time"

	"github.com/michaelquigley/df"
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
	SigningKey           string       `df:",secret"`
	EncryptionKey        string       `df:",secret"`
	Providers            []df.Dynamic `df:",secret"`
}

type OauthProviderConfig struct {
	Name         string
	ClientId     string
	ClientSecret string `df:",secret"`
}

func DefaultConfig() *Config {
	return &Config{
		Identity: "public",
		Address:  "0.0.0.0:8080",
	}
}

func (c *Config) Load(path string) error {
	opts := &df.Options{
		DynamicBinders: map[string]func(map[string]any) (df.Dynamic, error){
			(&githubConfig{}).Type(): newGithubConfig,
			(&googleConfig{}).Type(): newGoogleConfig,
			(&oidcConfig{}).Type():   newOidcConfig,
		},
	}
	if err := df.MergeFromYAML(c, path, opts); err != nil {
		return errors.Wrapf(err, "error loading frontend config '%v'", path)
	}
	if c.V != V {
		return errors.Errorf("invalid configuration version '%d'; expected '%d'", c.V, V)
	}
	return nil
}

func configureOauth(ctx context.Context, cfg *Config, tls bool) error {
	if cfg.Oauth == nil {
		logrus.Info("no oauth configuration; skipping oauth handler startup")
		return nil
	}

	for _, v := range cfg.Oauth.Providers {
		if prvCfg, ok := v.(df.Dynamic); ok {
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
