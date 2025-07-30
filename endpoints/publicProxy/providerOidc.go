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
	"github.com/openziti/zrok/endpoints/proxyUi"
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

	signingKey, err := deriveKey(c.cfg.SigningKey, 32)
	if err != nil {
		return err
	}
	encryptionKey, err := deriveKey(c.cfg.EncryptionKey, 32)
	if err != nil {
		return err
	}
	cookieHandler := zhttp.NewCookieHandler(signingKey, encryptionKey, zhttp.WithUnsecure(), zhttp.WithDomain(c.cfg.CookieDomain))
	redirectUrl := fmt.Sprintf("%v/%v/auth/callback", c.cfg.EndpointUrl, c.oidcCfg.Name)
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

	auth := func(w http.ResponseWriter, r *http.Request) {
		targetHost, err := url.QueryUnescape(r.URL.Query().Get("targetHost"))
		if err != nil {
			logrus.Errorf("unable to unescape 'targetHost': %v", err)
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
			s, err := t.SignedString(signingKey)
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
		logrus.Infof("invoking auth handler")
		rp.AuthURLHandler(state, provider, urlOptions...).ServeHTTP(w, r)
	}
	http.HandleFunc(fmt.Sprintf("/%v/login", c.oidcCfg.Name), auth)

	refresh := func(w http.ResponseWriter, r *http.Request) {
		scheme := "http"
		if c.tls {
			scheme = "https"
		}

		targetHost, err := url.QueryUnescape(r.URL.Query().Get("targetHost"))
		if err != nil {
			logrus.Errorf("unable to unescape 'targetHost': %v", err)
			proxyUi.WriteUnauthorized(w)
			return
		}

		logrus.Infof("refreshing for '%v'", targetHost)

		cookie, err := r.Cookie(c.cfg.CookieName)
		if err != nil {
			logrus.Errorf("unable to get 'zrok-access' cookie: %v", err)
			proxyUi.WriteUnauthorized(w)
			return
		}

		tkn, err := jwt.ParseWithClaims(cookie.Value, &zrokClaims{}, func(t *jwt.Token) (interface{}, error) {
			return signingKey, nil
		})
		if err != nil {
			logrus.Errorf("unable to parse jwt: %v", err)
			proxyUi.WriteUnauthorized(w)
			return
		}

		claims := tkn.Claims.(*zrokClaims)
		if claims.Provider != c.oidcCfg.Name {
			logrus.Error("token validation failed")
			proxyUi.WriteUnauthorized(w)
			return
		}

		accessToken, err := decryptToken(claims.AccessToken, encryptionKey)
		if err != nil {
			logrus.Errorf("unable to decrypt access token: %v", err)
			proxyUi.WriteUnauthorized(w)
			return
		}

		newTokens, err := rp.RefreshTokens[*oidc.IDTokenClaims](context.Background(), provider, accessToken, "", "")
		if err != nil {
			logrus.Errorf("unable to refresh tokens: %v", err)
			proxyUi.WriteUnauthorized(w)
			return
		}

		setSessionCookie(w, c.cfg, true, claims.Email, newTokens.AccessToken, c.oidcCfg.Name, claims.AuthorizationCheckInterval, signingKey, encryptionKey, targetHost)
		http.Redirect(w, r, fmt.Sprintf("%v://%v", scheme, targetHost), http.StatusFound)
	}
	http.HandleFunc(fmt.Sprintf("/%v/refresh", c.oidcCfg.Name), refresh)

	login := func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens[*oidc.IDTokenClaims], state string, provider rp.RelyingParty, info *oidc.UserInfo) {
		token, err := jwt.ParseWithClaims(state, &IntermediateJWT{}, func(t *jwt.Token) (interface{}, error) {
			return signingKey, nil
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
		setSessionCookie(w, c.cfg, true, info.Email, tokens.AccessToken, c.oidcCfg.Name, authCheckInterval, signingKey, encryptionKey, token.Claims.(*IntermediateJWT).Host)
		http.Redirect(w, r, fmt.Sprintf("%s://%s", scheme, token.Claims.(*IntermediateJWT).Host), http.StatusFound)
	}
	http.Handle(fmt.Sprintf("/%v/auth/callback", c.oidcCfg.Name), rp.CodeExchangeHandler(rp.UserinfoCallback(login), provider))

	logrus.Infof("configured oidc provider at '/%v'", c.oidcCfg.Name)

	return nil
}
