package dynamicProxy

import (
	"context"
	"net/http"
	"time"

	"github.com/michaelquigley/df"
	"github.com/openziti/zrok/endpoints"
	"github.com/openziti/zrok/endpoints/proxyUi"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	zhttp "github.com/zitadel/oidc/v3/pkg/http"
)

type config struct {
	V                      int    `df:"+match=1"`
	FrontendToken          string `df:"+required"`
	Identity               string
	BindAddress            string
	TemplatePath           string
	MappingRefreshInterval time.Duration
	Interstitial           *interstitialConfig
	Oauth                  *oauthConfig
	AmqpSubscriber         *amqpSubscriberConfig   `df:"+required"`
	Controller             *controllerClientConfig `df:"+required"`
	Tls                    *endpoints.TlsConfig
}

type interstitialConfig struct {
	Enabled           bool
	HtmlPath          string
	UserAgentPrefixes []string
}

type oauthConfig struct {
	BindAddress          string
	EndpointUrl          string
	CookieName           string
	CookieDomain         string
	SessionLifetime      time.Duration
	IntermediateLifetime time.Duration
	SigningKey           string `df:"+secret"`
	EncryptionKey        string `df:"+secret"`
	Providers            []df.Dynamic
}

type oauthProviderConfig struct {
	Name         string
	ClientId     string
	ClientSecret string `df:"+secret"`
}

func defaults() *config {
	return &config{
		Identity:               "public",
		BindAddress:            "0.0.0.0:8080",
		MappingRefreshInterval: 5 * time.Minute,
		AmqpSubscriber: &amqpSubscriberConfig{
			QueueDepth: 1024,
		},
		Controller: &controllerClientConfig{
			Timeout: 30 * time.Second,
		},
	}
}

func configureOauth(ctx context.Context, cfg *config, tls bool) error {
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
