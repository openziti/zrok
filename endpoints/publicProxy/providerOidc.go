package publicProxy

import (
	"context"
	"errors"
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
	cfg     *Config
	oidcCfg *oidcConfig
	tls     bool
}

func newOidcConfigurer(cfg *Config, tls bool, v map[string]interface{}) (*oidcConfigurer, error) {
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
	DiscoveryURL string   `mapstructure:"discovery_url"`
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

	signingKey, err := deriveKey(c.cfg.Oauth.SigningKey, 32)
	if err != nil {
		return err
	}
	encryptionKey, err := deriveKey(c.cfg.Oauth.EncryptionKey, 32)
	if err != nil {
		return err
	}
	cookieHandler := zhttp.NewCookieHandler(signingKey, encryptionKey, zhttp.WithUnsecure(), zhttp.WithDomain(c.cfg.Oauth.CookieDomain))
	redirectUrl := fmt.Sprintf("%v/%v/auth/callback", c.cfg.Oauth.EndpointUrl, c.oidcCfg.Name)
	providerOptions := []rp.Option{
		rp.WithCookieHandler(cookieHandler),
		rp.WithVerifierOpts(rp.WithIssuedAtOffset(5 * time.Second)),
	}
	if c.oidcCfg.DiscoveryURL != "" {
		rp.WithCustomDiscoveryUrl(c.oidcCfg.DiscoveryURL)
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
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("unable to unescape targetHost")), c.cfg.TemplatePath)
			return
		}
		state := func() string {
			id := uuid.New().String()
			t := jwt.NewWithClaims(jwt.SigningMethodHS256, IntermediateJWT{
				State:           id,
				TargetHost:      targetHost,
				RefreshInterval: r.URL.Query().Get("refreshInterval"),
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(c.cfg.Oauth.IntermediateLifetime)),
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
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("unable to unescape targetHost")), c.cfg.TemplatePath)
			return
		}

		cookie, err := r.Cookie(c.cfg.Oauth.CookieName)
		if err != nil {
			logrus.Errorf("unable to get auth session cookie: %v", err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("unable to get auth session cookie")), c.cfg.TemplatePath)
			return
		}

		tkn, err := jwt.ParseWithClaims(cookie.Value, &zrokClaims{}, func(t *jwt.Token) (interface{}, error) {
			return signingKey, nil
		})
		if err != nil {
			logrus.Errorf("unable to parse jwt: %v", err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("unable to parse jwt")), c.cfg.TemplatePath)
			return
		}

		claims := tkn.Claims.(*zrokClaims)
		if claims.Provider != c.oidcCfg.Name {
			logrus.Error("token provider mismatch")
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("token provider mismatch")), c.cfg.TemplatePath)
			return
		}

		accessToken, err := decryptToken(claims.AccessToken, encryptionKey)
		if err != nil {
			logrus.Errorf("unable to decrypt access token: %v", err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(errors.New("unable to decrypt access token")), c.cfg.TemplatePath)
			return
		}

		newTokens, err := rp.RefreshTokens[*oidc.IDTokenClaims](context.Background(), provider, accessToken, "", "")
		if err != nil {
			logrus.Errorf("unable to refresh tokens: %v", err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(errors.New("unable to refresh tokens")), c.cfg.TemplatePath)
			return
		}

		setSessionCookie(w, sessionCookieRequest{
			oauthCfg:        c.cfg.Oauth,
			supportsRefresh: true,
			email:           claims.Email,
			accessToken:     newTokens.AccessToken,
			provider:        c.oidcCfg.Name,
			refreshInterval: claims.RefreshInterval,
			signingKey:      signingKey,
			encryptionKey:   encryptionKey,
			targetHost:      targetHost,
		})

		http.Redirect(w, r, fmt.Sprintf("%v://%v", scheme, targetHost), http.StatusFound)
	}
	http.HandleFunc(fmt.Sprintf("/%v/refresh", c.oidcCfg.Name), refresh)

	login := func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens[*oidc.IDTokenClaims], state string, provider rp.RelyingParty, info *oidc.UserInfo) {
		token, err := jwt.ParseWithClaims(state, &IntermediateJWT{}, func(t *jwt.Token) (interface{}, error) {
			return signingKey, nil
		})
		if err != nil {
			logrus.Errorf("unable to parse intermediate JWT: %v", err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("unable to parse intermediate jwt")), c.cfg.TemplatePath)
			return
		}

		var refreshInterval time.Duration
		if v, err := time.ParseDuration(token.Claims.(*IntermediateJWT).RefreshInterval); err == nil {
			refreshInterval = v
		} else {
			logrus.Errorf("unable to parse authorization check interval: %v", err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(info.Email).WithError(errors.New("unable to parse authorization check interval")), c.cfg.TemplatePath)
			return
		}

		setSessionCookie(w, sessionCookieRequest{
			oauthCfg:        c.cfg.Oauth,
			supportsRefresh: true,
			email:           info.Email,
			accessToken:     tokens.AccessToken,
			provider:        c.oidcCfg.Name,
			refreshInterval: refreshInterval,
			signingKey:      signingKey,
			encryptionKey:   encryptionKey,
			targetHost:      token.Claims.(*IntermediateJWT).TargetHost,
		})

		http.Redirect(w, r, fmt.Sprintf("%s://%s", scheme, token.Claims.(*IntermediateJWT).TargetHost), http.StatusFound)
	}
	http.Handle(fmt.Sprintf("/%v/auth/callback", c.oidcCfg.Name), rp.CodeExchangeHandler(rp.UserinfoCallback(login), provider))

	logout := func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(c.cfg.Oauth.CookieName)
		if err == nil {
			tkn, err := jwt.ParseWithClaims(cookie.Value, &zrokClaims{}, func(t *jwt.Token) (interface{}, error) {
				return signingKey, nil
			})
			if err == nil {
				claims := tkn.Claims.(*zrokClaims)
				if claims.Provider == c.oidcCfg.Name {
					accessToken, err := decryptToken(claims.AccessToken, encryptionKey)
					if err == nil {
						if err := rp.RevokeToken(context.Background(), provider, accessToken, "access_token"); err == nil {
							logrus.Infof("revoked access token for '%v'", claims.Email)
						} else {
							logrus.Errorf("access token revocation failed: %v", err)
							proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(errors.New("access token revocation failed")), c.cfg.TemplatePath)
							return
						}
					} else {
						logrus.Errorf("unable to decrypt access token for '%v': %v", claims.Email, err)
						proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(errors.New("unable to decrypt access token")), c.cfg.TemplatePath)
						return
					}
				} else {
					logrus.Errorf("expected provider name '%v' got '%v'", c.oidcCfg.Name, claims.Email)
					proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(errors.New("provider mismatch")), c.cfg.TemplatePath)
					return
				}
			} else {
				logrus.Errorf("invalid jwt; unable to parse: %v", err)
				proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("invalid jwt; unable to parse")), c.cfg.TemplatePath)
				return
			}
		} else {
			logrus.Errorf("error getting cookie '%v': %v", c.cfg.Oauth.CookieName, err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("invalid cookie")), c.cfg.TemplatePath)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     c.cfg.Oauth.CookieName,
			Value:    "",
			MaxAge:   -1,
			Domain:   c.cfg.Oauth.CookieDomain,
			Path:     "/",
			HttpOnly: true,
		})

		redirectURL := r.URL.Query().Get("redirect_url")
		if redirectURL == "" {
			redirectURL = fmt.Sprintf("%s/%s/login", c.cfg.Oauth.EndpointUrl, c.oidcCfg.Name)
		}
		http.Redirect(w, r, redirectURL, http.StatusFound)
	}
	http.HandleFunc(fmt.Sprintf("/%v/logout", c.oidcCfg.Name), logout)

	logrus.Infof("configured oidc provider at '/%v'", c.oidcCfg.Name)

	return nil
}
