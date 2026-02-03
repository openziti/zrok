package publicProxy

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/openziti/zrok/v2/endpoints"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/michaelquigley/df/dd"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/endpoints/proxyUi"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	zhttp "github.com/zitadel/oidc/v3/pkg/http"
	"github.com/zitadel/oidc/v3/pkg/oidc"
)

type oidcConfig struct {
	Name         string
	ClientId     string
	ClientSecret string
	Scopes       []string
	Issuer       string
	DiscoveryURL string
	Pkce         bool
}

func newOidcConfig(v map[string]interface{}) (dd.Dynamic, error) {
	return dd.New[oidcConfig](v)
}

func (c *oidcConfig) Type() string                   { return "oidc" }
func (c *oidcConfig) ToMap() (map[string]any, error) { return nil, nil }

func (c *oidcConfig) configure(cfg *OauthConfig, tls bool) error {
	scheme := "http"
	if tls {
		scheme = "https"
	}

	signingKey, err := endpoints.DeriveKey(cfg.SigningKey, 32)
	if err != nil {
		return err
	}
	encryptionKey, err := endpoints.DeriveKey(cfg.EncryptionKey, 32)
	if err != nil {
		return err
	}
	cookieHandler := zhttp.NewCookieHandler(signingKey, encryptionKey, zhttp.WithUnsecure(), zhttp.WithDomain(cfg.CookieDomain))
	redirectUrl := fmt.Sprintf("%v/%v/auth/callback", cfg.EndpointUrl, c.Name)
	providerOptions := []rp.Option{
		rp.WithCookieHandler(cookieHandler),
		rp.WithVerifierOpts(rp.WithIssuedAtOffset(5 * time.Second)),
	}
	if c.DiscoveryURL != "" {
		rp.WithCustomDiscoveryUrl(c.DiscoveryURL)
	}
	provider, err := rp.NewRelyingPartyOIDC(
		context.TODO(),
		c.Issuer,
		c.ClientId,
		c.ClientSecret,
		redirectUrl,
		c.Scopes,
		providerOptions...,
	)
	if err != nil {
		return err
	}

	auth := func(w http.ResponseWriter, r *http.Request) {
		targetHost, err := url.QueryUnescape(r.URL.Query().Get("targetHost"))
		if err != nil {
			dl.Errorf("unable to unescape 'targetHost': %v", err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("unable to unescape targetHost")))
			return
		}
		state := func() string {
			id := uuid.New().String()
			t := jwt.NewWithClaims(jwt.SigningMethodHS256, IntermediateJWT{
				State:           id,
				TargetHost:      targetHost,
				RefreshInterval: r.URL.Query().Get("refreshInterval"),
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.IntermediateLifetime)),
					IssuedAt:  jwt.NewNumericDate(time.Now()),
					NotBefore: jwt.NewNumericDate(time.Now()),
					Issuer:    "zrok",
					Subject:   "intermediate_token",
					ID:        id,
				},
			})
			s, err := t.SignedString(signingKey)
			if err != nil {
				dl.Errorf("unable to sign intermediate JWT: %v", err)
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
	http.HandleFunc(fmt.Sprintf("/%v/login", c.Name), auth)

	refresh := func(w http.ResponseWriter, r *http.Request) {
		scheme := "http"
		if tls {
			scheme = "https"
		}

		targetHost, err := url.QueryUnescape(r.URL.Query().Get("targetHost"))
		if err != nil {
			dl.Errorf("unable to unescape 'targetHost': %v", err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("unable to unescape targetHost")))
			return
		}

		cookie, err := getSessionCookie(r, cfg.CookieName)
		if err != nil {
			dl.Errorf("unable to get auth session cookie: %v", err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("unable to get auth session cookie")))
			return
		}

		tkn, err := jwt.ParseWithClaims(cookie.Value, &zrokClaims{}, func(t *jwt.Token) (interface{}, error) {
			return signingKey, nil
		})
		if err != nil {
			dl.Errorf("unable to parse jwt: %v", err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("unable to parse jwt")))
			return
		}

		claims := tkn.Claims.(*zrokClaims)
		if claims.Provider != c.Name {
			dl.Error("token provider mismatch")
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("token provider mismatch")))
			return
		}

		accessToken, err := endpoints.DecryptToken(claims.AccessToken, encryptionKey)
		if err != nil {
			dl.Errorf("unable to decrypt access token: %v", err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(errors.New("unable to decrypt access token")))
			return
		}

		newTokens, err := rp.RefreshTokens[*oidc.IDTokenClaims](context.Background(), provider, accessToken, "", "")
		if err != nil {
			dl.Errorf("unable to refresh tokens: %v", err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(errors.New("unable to refresh tokens")))
			return
		}

		setSessionCookie(w, sessionCookieRequest{
			oauthCfg:        cfg,
			supportsRefresh: true,
			email:           claims.Email,
			accessToken:     newTokens.AccessToken,
			provider:        c.Name,
			refreshInterval: claims.RefreshInterval,
			signingKey:      signingKey,
			encryptionKey:   encryptionKey,
			targetHost:      targetHost,
		})

		http.Redirect(w, r, fmt.Sprintf("%v://%v", scheme, targetHost), http.StatusFound)
	}
	http.HandleFunc(fmt.Sprintf("/%v/refresh", c.Name), refresh)

	login := func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens[*oidc.IDTokenClaims], state string, provider rp.RelyingParty, info *oidc.UserInfo) {
		token, err := jwt.ParseWithClaims(state, &IntermediateJWT{}, func(t *jwt.Token) (interface{}, error) {
			return signingKey, nil
		})
		if err != nil {
			dl.Errorf("unable to parse intermediate JWT: %v", err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("unable to parse intermediate jwt")))
			return
		}

		var refreshInterval time.Duration
		if v, err := time.ParseDuration(token.Claims.(*IntermediateJWT).RefreshInterval); err == nil {
			refreshInterval = v
		} else {
			dl.Errorf("unable to parse authorization check interval: %v", err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(info.Email).WithError(errors.New("unable to parse authorization check interval")))
			return
		}

		setSessionCookie(w, sessionCookieRequest{
			oauthCfg:        cfg,
			supportsRefresh: true,
			email:           info.Email,
			accessToken:     tokens.AccessToken,
			provider:        c.Name,
			refreshInterval: refreshInterval,
			signingKey:      signingKey,
			encryptionKey:   encryptionKey,
			targetHost:      token.Claims.(*IntermediateJWT).TargetHost,
		})

		http.Redirect(w, r, fmt.Sprintf("%s://%s", scheme, token.Claims.(*IntermediateJWT).TargetHost), http.StatusFound)
	}
	http.Handle(fmt.Sprintf("/%v/auth/callback", c.Name), rp.CodeExchangeHandler(rp.UserinfoCallback(login), provider))

	logout := func(w http.ResponseWriter, r *http.Request) {
		cookie, err := getSessionCookie(r, cfg.CookieName)
		if err == nil {
			tkn, err := jwt.ParseWithClaims(cookie.Value, &zrokClaims{}, func(t *jwt.Token) (interface{}, error) {
				return signingKey, nil
			})
			if err == nil {
				claims := tkn.Claims.(*zrokClaims)
				if claims.Provider == c.Name {
					accessToken, err := endpoints.DecryptToken(claims.AccessToken, encryptionKey)
					if err == nil {
						if err := rp.RevokeToken(context.Background(), provider, accessToken, "access_token"); err == nil {
							dl.Infof("revoked access token for '%v'", claims.Email)
						} else {
							dl.Errorf("access token revocation failed: %v", err)
							proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(errors.New("access token revocation failed")))
							return
						}
					} else {
						dl.Errorf("unable to decrypt access token for '%v': %v", claims.Email, err)
						proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(errors.New("unable to decrypt access token")))
						return
					}
				} else {
					dl.Errorf("expected provider name '%v' got '%v'", c.Name, claims.Provider)
					proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(errors.New("provider mismatch")))
					return
				}
			} else {
				dl.Errorf("invalid jwt; unable to parse: %v", err)
				proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("invalid jwt; unable to parse")))
				return
			}
		} else {
			dl.Errorf("error getting cookie '%v': %v", cfg.CookieName, err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("invalid cookie")))
			return
		}

		clearSessionCookies(w, r, cfg.CookieName, cfg)

		redirectURL := r.URL.Query().Get("redirect_url")
		if redirectURL == "" {
			redirectURL = fmt.Sprintf("%s/%s/login", cfg.EndpointUrl, c.Name)
		}
		http.Redirect(w, r, redirectURL, http.StatusFound)
	}
	http.HandleFunc(fmt.Sprintf("/%v/logout", c.Name), logout)

	dl.Infof("configured oidc provider at '/%v'", c.Name)

	return nil
}
