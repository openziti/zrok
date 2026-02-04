package publicProxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/openziti/zrok/v2/endpoints"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/michaelquigley/df/dd"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/endpoints/proxyUi"
	"github.com/zitadel/oidc/v2/pkg/client/rp"
	zhttp "github.com/zitadel/oidc/v2/pkg/http"
	"github.com/zitadel/oidc/v2/pkg/oidc"
	"golang.org/x/oauth2"
	googleOauth "golang.org/x/oauth2/google"
)

type googleConfig struct {
	Name         string
	ClientId     string
	ClientSecret string
}

func newGoogleConfig(v map[string]interface{}) (dd.Dynamic, error) {
	return dd.New[googleConfig](v)
}

func (c *googleConfig) Type() string                   { return "google" }
func (c *googleConfig) ToMap() (map[string]any, error) { return nil, nil }

func (c *googleConfig) configure(cfg *OauthConfig, tls bool) error {
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
	rpConfig := &oauth2.Config{
		ClientID:     c.ClientId,
		ClientSecret: c.ClientSecret,
		RedirectURL:  fmt.Sprintf("%v/%v/auth/callback", cfg.EndpointUrl, c.Name),
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
		return err
	}

	type googleOauthEmailResp struct {
		Email string
	}

	auth := func(provider rp.RelyingParty) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			targetHost, err := url.QueryUnescape(r.URL.Query().Get("targetHost"))
			if err != nil {
				dl.Errorf("unable to unescape targetHost: %v", err)
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
			}, provider, rp.WithURLParam("access_type", "offline"), rp.URLParamOpt(rp.WithPrompt("login")))(w, r)
		}
	}
	http.Handle(fmt.Sprintf("/%v/login", c.Name), auth(provider))

	login := func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens[*oidc.IDTokenClaims], state string, rp rp.RelyingParty) {
		token, err := jwt.ParseWithClaims(state, &IntermediateJWT{}, func(t *jwt.Token) (interface{}, error) {
			return signingKey, nil
		})
		if err != nil {
			dl.Errorf("error parsing intermediate token: %v", err.Error())
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("error parsing intermediate token")))
			return
		}

		var refreshInterval time.Duration
		if v, err := time.ParseDuration(token.Claims.(*IntermediateJWT).RefreshInterval); err == nil {
			refreshInterval = v
		} else {
			dl.Errorf("unable to parse authorization check interval: %v", err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("unable to parse authorization check interval")))
			return
		}

		resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + url.QueryEscape(tokens.AccessToken))
		if err != nil {
			dl.Errorf("error getting user info from google: %v", err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("error getting user info from google")))
			return
		}
		defer func() {
			_ = resp.Body.Close()
		}()
		response, err := io.ReadAll(resp.Body)
		if err != nil {
			dl.Errorf("error reading response body: %v", err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("error reading google response body")))
			return
		}
		dl.Debugf("response from google userinfo endpoint: %s", string(response))
		data := googleOauthEmailResp{}
		err = json.Unmarshal(response, &data)
		if err != nil {
			dl.Errorf("error unmarshalling google oauth response: %v", err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("error unmarshalling google oauth response")))
			return
		}

		setSessionCookie(w, sessionCookieRequest{
			oauthCfg:        cfg,
			supportsRefresh: false,
			email:           data.Email,
			accessToken:     tokens.AccessToken,
			provider:        c.Name,
			refreshInterval: refreshInterval,
			signingKey:      signingKey,
			encryptionKey:   encryptionKey,
			targetHost:      token.Claims.(*IntermediateJWT).TargetHost,
		})

		http.Redirect(w, r, fmt.Sprintf("%s://%s", scheme, token.Claims.(*IntermediateJWT).TargetHost), http.StatusFound)
	}
	http.Handle(fmt.Sprintf("/%v/auth/callback", c.Name), rp.CodeExchangeHandler(login, provider))

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
						revokeURL := "https://oauth2.googleapis.com/revoke"
						resp, err := http.PostForm(revokeURL, url.Values{
							"token": {accessToken},
						})
						if err == nil {
							defer resp.Body.Close()
							if resp.StatusCode == http.StatusOK {
								dl.Infof("revoked google token for '%v'", claims.Email)
							} else {
								dl.Errorf("access token revocation failed with status: %v", resp.StatusCode)
								proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(errors.New("access token revocation failed")))
								return
							}
						} else {
							dl.Errorf("unable to revoke access token for '%v': %v", claims.Email, err)
							proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(errors.New("unable to post access token revocation")))
							return
						}
					} else {
						dl.Errorf("unable to decrypt access token for '%v': %v", claims.Email, err)
						proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(errors.New("unable to decrypt access token")))
						return
					}
				} else {
					dl.Errorf("expected provider name '%v' got '%v'", c.Name, claims.Provider)
					proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(errors.New("provider name mismatch")))
					return
				}
			} else {
				dl.Errorf("invalid jwt; unable to parse: %v", err)
				proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("invalid jwt; unable to parse")))
				return
			}
		} else {
			dl.Errorf("error getting cookie '%v': %v", cfg.CookieName, err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("error getting cookie")))
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

	dl.Infof("configured google provider at '/%v'", c.Name)

	return nil
}
