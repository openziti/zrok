package dynamicProxy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/michaelquigley/df/dd"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/endpoints/proxyUi"
	"github.com/pkg/errors"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	zhttp "github.com/zitadel/oidc/v3/pkg/http"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"golang.org/x/oauth2"
	githubOAuth "golang.org/x/oauth2/github"
)

func init() {
	registerOauthBinder((&githubConfig{}).Type(), newGithubConfig)
}

type githubConfig struct {
	Name         string
	ClientId     string
	ClientSecret string
}

func newGithubConfig(v map[string]interface{}) (dd.Dynamic, error) {
	return dd.New[githubConfig](v)
}

func (c *githubConfig) Type() string                   { return "github" }
func (c *githubConfig) ToMap() (map[string]any, error) { return nil, nil }

func (c *githubConfig) configure(cfg *oauthConfig, tls bool) error {
	// create github provider instance
	provider, err := createGithubProvider(c, cfg, tls)
	if err != nil {
		return err
	}

	// register with the oauth router
	return registerOAuthProvider(provider)
}

// githubProvider implements the oauthProvider interface for GitHub OAuth
type githubProvider struct {
	config        *githubConfig
	oauthCfg      *oauthConfig
	provider      rp.RelyingParty
	signingKey    []byte
	encryptionKey []byte
	tls           bool
}

// githubUserResp represents the response from GitHub's user emails endpoint
type githubUserResp struct {
	Email      string `json:"email"`
	Primary    bool   `json:"primary"`
	Verified   bool   `json:"verified"`
	Visibility string `json:"visibility"`
}

// Name returns the provider name
func (p *githubProvider) Name() string {
	return p.config.Name
}

// RegisterRoutes registers the GitHub OAuth routes with the provided router
func (p *githubProvider) RegisterRoutes(router *mux.Router) error {
	// register login route
	router.Handle(fmt.Sprintf("/%v/login", p.config.Name), p.authHandler())

	// register callback route
	router.Handle(fmt.Sprintf("/%v/auth/callback", p.config.Name), rp.CodeExchangeHandler(p.loginHandler(), p.provider))

	// register logout route
	router.HandleFunc(fmt.Sprintf("/%v/logout", p.config.Name), p.logoutHandler())

	dl.Debugf("registered github provider routes at '/%v'", p.config.Name)
	return nil
}

// authHandler creates the authentication handler for initiating OAuth flow
func (p *githubProvider) authHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetHost, err := url.QueryUnescape(r.URL.Query().Get("targetHost"))
		if err != nil {
			err := fmt.Errorf("unable to unescape targetHost: %v", err)
			dl.Error(err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(err))
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
				dl.Errorf("unable to sign intermediate JWT: %v", err)
			}
			return s
		}, p.provider, rp.WithURLParam("access_type", "offline"), rp.URLParamOpt(rp.WithPrompt("login")))(w, r)
	})
}

// loginHandler creates the login callback handler for processing OAuth responses
func (p *githubProvider) loginHandler() func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens[*oidc.IDTokenClaims], state string, rp rp.RelyingParty) {
	return func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens[*oidc.IDTokenClaims], state string, rp rp.RelyingParty) {
		token, err := jwt.ParseWithClaims(state, &IntermediateJWT{}, func(t *jwt.Token) (interface{}, error) {
			return p.signingKey, nil
		})
		if err != nil {
			errOut := errors.Wrap(err, "error parsing intermediate token")
			dl.Error(errOut)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errOut))
			return
		}

		var refreshInterval time.Duration
		if v, err := time.ParseDuration(token.Claims.(*IntermediateJWT).RefreshInterval); err == nil {
			refreshInterval = v
		} else {
			errOut := errors.Wrapf(err, "unable to parse authorization check interval '%v'", token.Claims.(*IntermediateJWT).RefreshInterval)
			dl.Error(errOut)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errOut))
			return
		}

		// get user emails from github
		parsedUrl, err := url.Parse("https://api.github.com/user/emails")
		if err != nil {
			errOut := errors.Wrap(err, "error parsing github url")
			dl.Error(errOut)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errOut))
			return
		}
		req := &http.Request{
			Method: http.MethodGet,
			URL:    parsedUrl,
			Header: make(http.Header),
		}
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokens.AccessToken))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			errOut := errors.Wrap(err, "error getting user info from github")
			dl.Error(errOut)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errOut))
			return
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		response, err := io.ReadAll(resp.Body)
		if err != nil {
			errOut := errors.Wrap(err, "error reading response body from github")
			dl.Error(errOut)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errOut))
			return
		}

		var rDat []githubUserResp
		err = json.Unmarshal(response, &rDat)
		if err != nil {
			errOut := errors.Wrap(err, "error unmarshalling response from github")
			dl.Error(errOut)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errOut))
			return
		}

		primaryEmail := ""
		for _, email := range rDat {
			if email.Primary {
				primaryEmail = email.Email
				break
			}
		}

		// set session cookie
		setSessionCookie(w, sessionCookieRequest{
			oauthCfg:        p.oauthCfg,
			supportsRefresh: false,
			email:           primaryEmail,
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

// logoutHandler creates the logout handler for revoking GitHub tokens and clearing cookies
func (p *githubProvider) logoutHandler() http.HandlerFunc {
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
						// revoke github token
						req, err := http.NewRequest("DELETE",
							fmt.Sprintf("https://api.github.com/applications/%s/token", p.config.ClientId),
							strings.NewReader(fmt.Sprintf(`{"access_token":"%s"}`, accessToken)))
						if err != nil {
							dl.Errorf("error creating access token delete request for '%v': %v", claims.Email, err)
							proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(errors.New("error creating access token delete request")))
							return
						}

						req.Header.Set("Content-Type", "application/json")
						req.SetBasicAuth(p.config.ClientId, p.config.ClientSecret) // Need client credentials

						resp, err := http.DefaultClient.Do(req)
						if err != nil {
							dl.Errorf("error invoking access token delete request for '%v': %v", claims.Email, err)
							proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(errors.New("error executing access token delete request")))
							return
						}
						defer resp.Body.Close()

						if resp.StatusCode == http.StatusNoContent {
							dl.Infof("revoked github access token for '%v'", claims.Email)
						} else {
							dl.Errorf("access token revocation failed with status: %v", resp.StatusCode)
							proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(errors.New("access token revocation failed")))
							return
						}
					} else {
						dl.Errorf("unable to decrypt access token for '%v': %v", claims.Email, err)
						proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(errors.New("unable to decrypt access token")))
						return
					}
				} else {
					dl.Errorf("expected provider name '%v' got '%v'", p.config.Name, claims.Provider)
					proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(errors.New("provider name mismatch")))
					return
				}
			} else {
				dl.Errorf("invalid jwt; unable to parse: %v", err)
				proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("invalid jwt; unable to parse")))
				return
			}
		} else {
			dl.Errorf("error getting cookie '%v': %v", p.oauthCfg.CookieName, err)
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

// createGithubProvider creates a new GitHub OAuth provider
func createGithubProvider(config *githubConfig, oauthCfg *oauthConfig, tls bool) (*githubProvider, error) {
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
		Scopes:       []string{"user:email"},
		Endpoint:     githubOAuth.Endpoint,
	}
	providerOptions := []rp.Option{
		rp.WithCookieHandler(cookieHandler),
		rp.WithVerifierOpts(rp.WithIssuedAtOffset(5 * time.Second)),
	}
	provider, err := rp.NewRelyingPartyOAuth(rpConfig, providerOptions...)
	if err != nil {
		return nil, err
	}

	return &githubProvider{
		config:        config,
		oauthCfg:      oauthCfg,
		provider:      provider,
		signingKey:    signingKey,
		encryptionKey: encryptionKey,
		tls:           tls,
	}, nil
}
