package dynamicProxy

import (
	"time"

	"github.com/michaelquigley/df"
	"github.com/openziti/zrok/endpoints"
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
