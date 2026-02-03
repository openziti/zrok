package dynamicProxy

import (
	"time"

	"github.com/michaelquigley/df/dd"
	"github.com/openziti/zrok/v2/endpoints"
)

type config struct {
	V                      int    `dd:"+match=1"`
	FrontendToken          string `dd:"+required"`
	Identity               string
	BindAddress            string
	TemplatePath           string
	MappingRefreshInterval time.Duration
	Interstitial           *interstitialConfig
	Oauth                  *oauthConfig
	AmqpSubscriber         *amqpSubscriberConfig   `dd:"+required"`
	Controller             *controllerClientConfig `dd:"+required"`
	Tls                    *endpoints.TlsConfig
}

type interstitialConfig struct {
	Enabled           bool
	HtmlPath          string
	UserAgentPrefixes []string
}

type oauthConfig struct {
	BindAddress          string `dd:"+required"`
	EndpointUrl          string `dd:"+required"`
	CookieName           string `dd:"+required"`
	CookieDomain         string `dd:"+required"`
	MaxCookieSize        int
	SessionLifetime      time.Duration `dd:"+required"`
	IntermediateLifetime time.Duration `dd:"+required"`
	SigningKey           string        `dd:"+secret"`
	EncryptionKey        string        `dd:"+secret"`
	Providers            []dd.Dynamic
}

func (c *oauthConfig) GetCookieName() string {
	return c.CookieName
}

func (c *oauthConfig) GetCookieDomain() string {
	return c.CookieDomain
}

func (c *oauthConfig) GetMaxCookieSize() int {
	return c.MaxCookieSize
}

func (c *oauthConfig) GetSessionLifetime() time.Duration {
	return c.SessionLifetime
}

type oauthProviderConfig struct {
	Name         string `dd:"+required"`
	ClientId     string `dd:"+required"`
	ClientSecret string `dd:"+secret,+required"`
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
