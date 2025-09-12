package dynamicProxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/michaelquigley/df"
	"github.com/openziti/zrok/endpoints/proxyUi"
	"github.com/sirupsen/logrus"
	"github.com/zitadel/oidc/v2/pkg/client/rp"
	zhttp "github.com/zitadel/oidc/v2/pkg/http"
	"github.com/zitadel/oidc/v2/pkg/oidc"
	"golang.org/x/oauth2"
	googleOauth "golang.org/x/oauth2/google"
)

func init() {
	registerOauthBinder((&googleConfig{}).Type(), newGoogleConfig)
}

type googleConfig struct {
	Name         string
	ClientId     string
	ClientSecret string
}

func newGoogleConfig(v map[string]interface{}) (df.Dynamic, error) {
	return df.New[googleConfig](v)
}

func (c *googleConfig) Type() string                   { return "google" }
func (c *googleConfig) ToMap() (map[string]any, error) { return nil, nil }

func (c *googleConfig) configure(cfg *oauthConfig, tls bool) error {
	// create google provider instance
	provider, err := createGoogleProvider(c, cfg, tls)
	if err != nil {
		return err
	}

	// register with the oauth router
	return registerOAuthProvider(provider)
}

// googleProvider implements the oauthProvider interface for Google OAuth
type googleProvider struct {
	config        *googleConfig
	oauthCfg      *oauthConfig
	provider      rp.RelyingParty
	signingKey    []byte
	encryptionKey []byte
	tls           bool
}

// googleOauthEmailResp represents the response from Google's userinfo endpoint
type googleOauthEmailResp struct {
	Email string `json:"email"`
}

// Name returns the provider name
func (p *googleProvider) Name() string {
	return p.config.Name
}

// RegisterRoutes registers the Google OAuth routes with the provided router
func (p *googleProvider) RegisterRoutes(router *mux.Router) error {
	// register login route
	router.Handle(fmt.Sprintf("/%v/login", p.config.Name), p.authHandler())

	// register callback route
	router.Handle(fmt.Sprintf("/%v/auth/callback", p.config.Name),
		rp.CodeExchangeHandler(p.loginHandler(), p.provider))

	// register logout route
	router.HandleFunc(fmt.Sprintf("/%v/logout", p.config.Name), p.logoutHandler())

	logrus.Debugf("registered google provider routes at '/%v'", p.config.Name)
	return nil
}

// authHandler creates the authentication handler for initiating OAuth flow
func (p *googleProvider) authHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetHost, err := url.QueryUnescape(r.URL.Query().Get("targetHost"))
		if err != nil {
			logrus.Errorf("unable to unescape targetHost: %v", err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("unable to escape targetHost")))
			return
		}

		rp.AuthURLHandler(func() string {
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
		}, p.provider, rp.WithURLParam("access_type", "offline"), rp.URLParamOpt(rp.WithPrompt("login")))(w, r)
	})
}

// loginHandler creates the login callback handler for processing OAuth responses
func (p *googleProvider) loginHandler() func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens[*oidc.IDTokenClaims], state string, rp rp.RelyingParty) {
	return func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens[*oidc.IDTokenClaims], state string, rp rp.RelyingParty) {
		token, err := jwt.ParseWithClaims(state, &IntermediateJWT{}, func(t *jwt.Token) (interface{}, error) {
			return p.signingKey, nil
		})
		if err != nil {
			logrus.Errorf("error parsing intermediate token: %v", err.Error())
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("error parsing intermediate token")))
			return
		}

		var refreshInterval time.Duration
		if v, err := time.ParseDuration(token.Claims.(*IntermediateJWT).RefreshInterval); err == nil {
			refreshInterval = v
		} else {
			logrus.Errorf("unable to parse authorization check interval: %v", err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("unable to parse authorization check interval")))
			return
		}

		// get user info from google
		resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + url.QueryEscape(tokens.AccessToken))
		if err != nil {
			logrus.Errorf("error getting user info from google: %v", err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("error getting user info from google")))
			return
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		response, err := io.ReadAll(resp.Body)
		if err != nil {
			logrus.Errorf("error reading response body: %v", err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("error reading google response body")))
			return
		}

		logrus.Debugf("response from google userinfo endpoint: %s", string(response))
		data := googleOauthEmailResp{}
		err = json.Unmarshal(response, &data)
		if err != nil {
			logrus.Errorf("error unmarshalling google oauth response: %v", err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("error unmarshalling google oauth response")))
			return
		}

		// set session cookie
		setSessionCookie(w, sessionCookieRequest{
			oauthCfg:        p.oauthCfg,
			supportsRefresh: false,
			email:           data.Email,
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

// logoutHandler creates the logout handler for revoking Google tokens and clearing cookies
func (p *googleProvider) logoutHandler() http.HandlerFunc {
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
						// revoke google token
						revokeURL := "https://oauth2.googleapis.com/revoke"
						resp, err := http.PostForm(revokeURL, url.Values{
							"token": {accessToken},
						})
						if err == nil {
							defer resp.Body.Close()
							if resp.StatusCode == http.StatusOK {
								logrus.Infof("revoked google token for '%v'", claims.Email)
							} else {
								logrus.Errorf("access token revocation failed with status: %v", resp.StatusCode)
								proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(errors.New("access token revocation failed")))
								return
							}
						} else {
							logrus.Errorf("unable to revoke access token for '%v': %v", claims.Email, err)
							proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(errors.New("unable to post access token revocation")))
							return
						}
					} else {
						logrus.Errorf("unable to decrypt access token for '%v': %v", claims.Email, err)
						proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(errors.New("unable to decrypt access token")))
						return
					}
				} else {
					logrus.Errorf("expected provider name '%v' got '%v'", p.config.Name, claims.Provider)
					proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(errors.New("provider name mismatch")))
					return
				}
			} else {
				logrus.Errorf("invalid jwt; unable to parse: %v", err)
				proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("invalid jwt; unable to parse")))
				return
			}
		} else {
			logrus.Errorf("error getting cookie '%v': %v", p.oauthCfg.CookieName, err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("error getting cookie")))
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

// createGoogleProvider creates a new Google OAuth provider
func createGoogleProvider(config *googleConfig, oauthCfg *oauthConfig, tls bool) (*googleProvider, error) {
	signingKey, err := deriveKey(oauthCfg.SigningKey, 32)
	if err != nil {
		return nil, err
	}
	encryptionKey, err := deriveKey(oauthCfg.EncryptionKey, 32)
	if err != nil {
		return nil, err
	}

	cookieHandler := zhttp.NewCookieHandler(signingKey, encryptionKey, zhttp.WithUnsecure(), zhttp.WithDomain(oauthCfg.CookieDomain))
	rpConfig := &oauth2.Config{
		ClientID:     config.ClientId,
		ClientSecret: config.ClientSecret,
		RedirectURL:  fmt.Sprintf("%v/%v/auth/callback", oauthCfg.EndpointUrl, config.Name),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     googleOauth.Endpoint,
	}
	providerOptions := []rp.Option{
		rp.WithCookieHandler(cookieHandler),
		rp.WithVerifierOpts(rp.WithIssuedAtOffset(5 * time.Second)),
		rp.WithPKCE(cookieHandler),
	}
	provider, err := rp.NewRelyingPartyOAuth(rpConfig, providerOptions...)
	if err != nil {
		return nil, err
	}

	return &googleProvider{
		config:        config,
		oauthCfg:      oauthCfg,
		provider:      provider,
		signingKey:    signingKey,
		encryptionKey: encryptionKey,
		tls:           tls,
	}, nil
}
