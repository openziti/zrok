package dynamicProxy

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/michaelquigley/df/dd"
	"github.com/openziti/zrok/endpoints/proxyUi"
	"github.com/sirupsen/logrus"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	zhttp "github.com/zitadel/oidc/v3/pkg/http"
	"github.com/zitadel/oidc/v3/pkg/oidc"
)

func init() {
	registerOauthBinder((&oidcConfig{}).Type(), newOidcConfig)
}

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

func (c *oidcConfig) configure(cfg *oauthConfig, tls bool) error {
	// create oidc provider instance
	provider, err := createOidcProvider(c, cfg, tls)
	if err != nil {
		return err
	}

	// register with the oauth router
	return registerOAuthProvider(provider)
}

// oidcProvider implements the oauthProvider interface for OIDC OAuth
type oidcProvider struct {
	config        *oidcConfig
	oauthCfg      *oauthConfig
	provider      rp.RelyingParty
	signingKey    []byte
	encryptionKey []byte
	tls           bool
}

// Name returns the provider name
func (p *oidcProvider) Name() string {
	return p.config.Name
}

// RegisterRoutes registers the OIDC OAuth routes with the provided router
func (p *oidcProvider) RegisterRoutes(router *mux.Router) error {
	// register login route
	router.HandleFunc(fmt.Sprintf("/%v/login", p.config.Name), p.authHandler())

	// register refresh route (unique to OIDC provider)
	router.HandleFunc(fmt.Sprintf("/%v/refresh", p.config.Name), p.refreshHandler())

	// register callback route
	router.Handle(fmt.Sprintf("/%v/auth/callback", p.config.Name),
		rp.CodeExchangeHandler(rp.UserinfoCallback(p.loginHandler()), p.provider))

	// register logout route
	router.HandleFunc(fmt.Sprintf("/%v/logout", p.config.Name), p.logoutHandler())

	logrus.Debugf("registered oidc provider routes at '/%v'", p.config.Name)
	return nil
}

// authHandler creates the authentication handler for initiating OAuth flow
func (p *oidcProvider) authHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		targetHost, err := url.QueryUnescape(r.URL.Query().Get("targetHost"))
		if err != nil {
			logrus.Errorf("unable to unescape 'targetHost': %v", err)
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
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(p.oauthCfg.IntermediateLifetime)),
					IssuedAt:  jwt.NewNumericDate(time.Now()),
					NotBefore: jwt.NewNumericDate(time.Now()),
					Issuer:    "zrok",
					Subject:   "intermediate_token",
					ID:        id,
				},
			})
			s, err := t.SignedString(p.signingKey)
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
		rp.AuthURLHandler(state, p.provider, urlOptions...).ServeHTTP(w, r)
	}
}

// refreshHandler creates the refresh handler for refreshing tokens
func (p *oidcProvider) refreshHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		scheme := "http"
		if p.tls {
			scheme = "https"
		}

		targetHost, err := url.QueryUnescape(r.URL.Query().Get("targetHost"))
		if err != nil {
			logrus.Errorf("unable to unescape 'targetHost': %v", err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("unable to unescape targetHost")))
			return
		}

		cookie, err := r.Cookie(p.oauthCfg.CookieName)
		if err != nil {
			logrus.Errorf("unable to get auth session cookie: %v", err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("unable to get auth session cookie")))
			return
		}

		tkn, err := jwt.ParseWithClaims(cookie.Value, &zrokClaims{}, func(t *jwt.Token) (interface{}, error) {
			return p.signingKey, nil
		})
		if err != nil {
			logrus.Errorf("unable to parse jwt: %v", err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("unable to parse jwt")))
			return
		}

		claims := tkn.Claims.(*zrokClaims)
		if claims.Provider != p.config.Name {
			logrus.Error("token provider mismatch")
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("token provider mismatch")))
			return
		}

		accessToken, err := decryptToken(claims.AccessToken, p.encryptionKey)
		if err != nil {
			logrus.Errorf("unable to decrypt access token: %v", err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(errors.New("unable to decrypt access token")))
			return
		}

		newTokens, err := rp.RefreshTokens[*oidc.IDTokenClaims](context.Background(), p.provider, accessToken, "", "")
		if err != nil {
			logrus.Errorf("unable to refresh tokens: %v", err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(errors.New("unable to refresh tokens")))
			return
		}

		setSessionCookie(w, sessionCookieRequest{
			oauthCfg:        p.oauthCfg,
			supportsRefresh: true,
			email:           claims.Email,
			accessToken:     newTokens.AccessToken,
			provider:        p.config.Name,
			refreshInterval: claims.RefreshInterval,
			signingKey:      p.signingKey,
			encryptionKey:   p.encryptionKey,
			targetHost:      targetHost,
		})

		http.Redirect(w, r, fmt.Sprintf("%v://%v", scheme, targetHost), http.StatusFound)
	}
}

// loginHandler creates the login callback handler for processing OAuth responses
func (p *oidcProvider) loginHandler() func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens[*oidc.IDTokenClaims], state string, provider rp.RelyingParty, info *oidc.UserInfo) {
	return func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens[*oidc.IDTokenClaims], state string, provider rp.RelyingParty, info *oidc.UserInfo) {
		token, err := jwt.ParseWithClaims(state, &IntermediateJWT{}, func(t *jwt.Token) (interface{}, error) {
			return p.signingKey, nil
		})
		if err != nil {
			logrus.Errorf("unable to parse intermediate JWT: %v", err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("unable to parse intermediate jwt")))
			return
		}

		var refreshInterval time.Duration
		if v, err := time.ParseDuration(token.Claims.(*IntermediateJWT).RefreshInterval); err == nil {
			refreshInterval = v
		} else {
			logrus.Errorf("unable to parse authorization check interval: %v", err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(info.Email).WithError(errors.New("unable to parse authorization check interval")))
			return
		}

		setSessionCookie(w, sessionCookieRequest{
			oauthCfg:        p.oauthCfg,
			supportsRefresh: true,
			email:           info.Email,
			accessToken:     tokens.AccessToken,
			provider:        p.config.Name,
			refreshInterval: refreshInterval,
			signingKey:      p.signingKey,
			encryptionKey:   p.encryptionKey,
			targetHost:      token.Claims.(*IntermediateJWT).TargetHost,
		})

		scheme := "http"
		if p.tls {
			scheme = "https"
		}
		http.Redirect(w, r, fmt.Sprintf("%s://%s", scheme, token.Claims.(*IntermediateJWT).TargetHost), http.StatusFound)
	}
}

// logoutHandler creates the logout handler for revoking OIDC tokens and clearing cookies
func (p *oidcProvider) logoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(p.oauthCfg.CookieName)
		if err == nil {
			tkn, err := jwt.ParseWithClaims(cookie.Value, &zrokClaims{}, func(t *jwt.Token) (interface{}, error) {
				return p.signingKey, nil
			})
			if err == nil {
				claims := tkn.Claims.(*zrokClaims)
				if claims.Provider == p.config.Name {
					accessToken, err := decryptToken(claims.AccessToken, p.encryptionKey)
					if err == nil {
						if err := rp.RevokeToken(context.Background(), p.provider, accessToken, "access_token"); err == nil {
							logrus.Infof("revoked access token for '%v'", claims.Email)
						} else {
							logrus.Errorf("access token revocation failed: %v", err)
							proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(errors.New("access token revocation failed")))
							return
						}
					} else {
						logrus.Errorf("unable to decrypt access token for '%v': %v", claims.Email, err)
						proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(errors.New("unable to decrypt access token")))
						return
					}
				} else {
					logrus.Errorf("expected provider name '%v' got '%v'", p.config.Name, claims.Provider)
					proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(errors.New("provider mismatch")))
					return
				}
			} else {
				logrus.Errorf("invalid jwt; unable to parse: %v", err)
				proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("invalid jwt; unable to parse")))
				return
			}
		} else {
			logrus.Errorf("error getting cookie '%v': %v", p.oauthCfg.CookieName, err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("invalid cookie")))
			return
		}

		// clear cookie
		http.SetCookie(w, &http.Cookie{
			Name:     p.oauthCfg.CookieName,
			Value:    "",
			MaxAge:   -1,
			Domain:   p.oauthCfg.CookieDomain,
			Path:     "/",
			HttpOnly: true,
		})

		redirectURL := r.URL.Query().Get("redirect_url")
		if redirectURL == "" {
			redirectURL = fmt.Sprintf("%s/%s/login", p.oauthCfg.EndpointUrl, p.config.Name)
		}
		http.Redirect(w, r, redirectURL, http.StatusFound)
	}
}

// createOidcProvider creates a new OIDC OAuth provider
func createOidcProvider(config *oidcConfig, oauthCfg *oauthConfig, tls bool) (*oidcProvider, error) {
	signingKey, err := deriveKey(oauthCfg.SigningKey, 32)
	if err != nil {
		return nil, err
	}
	encryptionKey, err := deriveKey(oauthCfg.EncryptionKey, 32)
	if err != nil {
		return nil, err
	}

	cookieHandler := zhttp.NewCookieHandler(signingKey, encryptionKey, zhttp.WithUnsecure(), zhttp.WithDomain(oauthCfg.CookieDomain))
	redirectUrl := fmt.Sprintf("%v/%v/auth/callback", oauthCfg.EndpointUrl, config.Name)
	providerOptions := []rp.Option{
		rp.WithCookieHandler(cookieHandler),
		rp.WithVerifierOpts(rp.WithIssuedAtOffset(5 * time.Second)),
	}
	if config.DiscoveryURL != "" {
		providerOptions = append(providerOptions, rp.WithCustomDiscoveryUrl(config.DiscoveryURL))
	}

	provider, err := rp.NewRelyingPartyOIDC(
		context.TODO(),
		config.Issuer,
		config.ClientId,
		config.ClientSecret,
		redirectUrl,
		config.Scopes,
		providerOptions...,
	)
	if err != nil {
		return nil, err
	}

	return &oidcProvider{
		config:        config,
		oauthCfg:      oauthCfg,
		provider:      provider,
		signingKey:    signingKey,
		encryptionKey: encryptionKey,
		tls:           tls,
	}, nil
}
