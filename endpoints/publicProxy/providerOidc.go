package publicProxy

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	zhttp "github.com/zitadel/oidc/v3/pkg/http"
	"github.com/zitadel/oidc/v3/pkg/oidc"
)

type oidcConfigurer struct {
	cfg     *OauthConfig
	oidcCfg *oidcConfig
	tls     bool
}

func newOidcConfigurer(cfg *OauthConfig, tls bool, v map[string]interface{}) (*oidcConfigurer, error) {
	c := &oidcConfigurer{cfg: cfg}
	oidcCfg, err := newOidcConfig(v)
	if err != nil {
		return nil, err
	}
	c.oidcCfg = oidcCfg
	c.tls = tls
	return c, nil
}

type oidcConfig struct {
	Name         string   `mapstructure:"name"`
	ClientId     string   `mapstructure:"client_id"`
	ClientSecret string   `mapstructure:"client_secret"`
	Scopes       []string `mapstructure:"scopes"`
	Issuer       string   `mapstructure:"issuer"`
	Pkce         bool     `mapstructure:"pkce"`
}

func newOidcConfig(v map[string]interface{}) (*oidcConfig, error) {
	cfg := &oidcConfig{}
	if err := mapstructure.Decode(v, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *oidcConfigurer) configure() error {
	scheme := "http"
	if c.tls {
		scheme = "https"
	}

	key, err := DeriveKey(c.cfg.HashKey, 32)
	if err != nil {
		return err
	}
	cookieHandler := zhttp.NewCookieHandler(key, key, zhttp.WithUnsecure(), zhttp.WithDomain(c.cfg.CookieDomain))
	redirectUrl := fmt.Sprintf("%v/%v/auth/callback", c.cfg.RedirectUrl, c.oidcCfg.Name)
	providerOptions := []rp.Option{
		rp.WithCookieHandler(cookieHandler),
		rp.WithVerifierOpts(rp.WithIssuedAtOffset(5 * time.Second)),
	}
	provider, err := rp.NewRelyingPartyOIDC(
		context.TODO(),
		c.oidcCfg.Issuer,
		c.oidcCfg.ClientId,
		c.oidcCfg.ClientSecret,
		redirectUrl,
		c.oidcCfg.Scopes,
		providerOptions...,
	)
	if err != nil {
		return err
	}

	authHandler := func(w http.ResponseWriter, r *http.Request) {
		targetHost, err := url.QueryUnescape(r.URL.Query().Get("targetHost"))
		if err != nil {
			logrus.Errorf("unable to unescape 'targethost': %v", err)
			http.Error(w, "invalid url format", http.StatusBadRequest)
			return
		}
		state := func() string {
			id := uuid.New().String()
			t := jwt.NewWithClaims(jwt.SigningMethodHS256, IntermediateJWT{
				id,
				targetHost,
				r.URL.Query().Get("checkInterval"),
				jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
					IssuedAt:  jwt.NewNumericDate(time.Now()),
					NotBefore: jwt.NewNumericDate(time.Now()),
					Issuer:    "zrok",
					Subject:   "intermediate_token",
					ID:        id,
				},
			})
			s, err := t.SignedString(key)
			if err != nil {
				logrus.Errorf("unable to sign intermediate JWT: %v", err)
			}
			return s
		}
		urlOptions := []rp.URLParamOpt{
			rp.WithPromptURLParam("login"),
			rp.WithResponseModeURLParam("query"),
			rp.WithURLParam("access_type", "offline"),
		}
		rp.AuthURLHandler(state, provider, urlOptions...)
	}
	http.HandleFunc(fmt.Sprintf("/%v/login", c.oidcCfg.Name), authHandler)

	login := func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens[*oidc.IDTokenClaims], state string, provider rp.RelyingParty, info *oidc.UserInfo) {
		token, err := jwt.ParseWithClaims(state, &IntermediateJWT{}, func(t *jwt.Token) (interface{}, error) {
			return key, nil
		})
		authCheckInterval := 3 * time.Hour
		i, err := time.ParseDuration(token.Claims.(*IntermediateJWT).AuthorizationCheckInterval)
		if err != nil {
			logrus.Errorf("unable to parse authorization check interval: %v. Defaulting to 3 hours", err)
		} else {
			authCheckInterval = i
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		SetZrokCookie(w, c.cfg.CookieDomain, info.Email, tokens.AccessToken, c.oidcCfg.Name, authCheckInterval, key, token.Claims.(*IntermediateJWT).Host)
		http.Redirect(w, r, fmt.Sprintf("%s://%s", scheme, token.Claims.(*IntermediateJWT).Host), http.StatusFound)
	}
	http.Handle(fmt.Sprintf("/%v/auth/callback", c.oidcCfg.Name), rp.CodeExchangeHandler(rp.UserinfoCallback(login), provider))

	logrus.Infof("configured oidc provider '%v'", c.oidcCfg.Name)

	return nil
}
