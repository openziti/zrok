package publicProxy

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
	"github.com/mitchellh/mapstructure"
	"github.com/openziti/zrok/endpoints/proxyUi"
	"github.com/sirupsen/logrus"
	"github.com/zitadel/oidc/v2/pkg/client/rp"
	zhttp "github.com/zitadel/oidc/v2/pkg/http"
	"github.com/zitadel/oidc/v2/pkg/oidc"
	"golang.org/x/oauth2"
	googleOauth "golang.org/x/oauth2/google"
)

type googleConfigurer struct {
	cfg       *OauthConfig
	googleCfg *googleConfig
	tls       bool
}

func newGoogleConfigurer(cfg *OauthConfig, tls bool, v map[string]interface{}) (*googleConfigurer, error) {
	c := &googleConfigurer{cfg: cfg}
	googleCfg, err := newGoogleConfig(v)
	if err != nil {
		return nil, err
	}
	c.googleCfg = googleCfg
	c.tls = tls
	return c, nil
}

type googleConfig struct {
	Name         string `mapstructure:"name"`
	ClientId     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
}

func newGoogleConfig(v map[string]interface{}) (*googleConfig, error) {
	cfg := &googleConfig{}
	if err := mapstructure.Decode(v, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *googleConfigurer) configure() error {
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
	rpConfig := &oauth2.Config{
		ClientID:     c.googleCfg.ClientId,
		ClientSecret: c.googleCfg.ClientSecret,
		RedirectURL:  fmt.Sprintf("%v/%v/auth/callback", c.cfg.EndpointUrl, c.googleCfg.Name),
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
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(c.cfg.IntermediateLifetime)),
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
			}, provider, rp.WithURLParam("access_type", "offline"), rp.URLParamOpt(rp.WithPrompt("login")))(w, r)
		}
	}
	http.Handle(fmt.Sprintf("/%v/login", c.googleCfg.Name), auth(provider))

	login := func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens[*oidc.IDTokenClaims], state string, rp rp.RelyingParty) {
		token, err := jwt.ParseWithClaims(state, &IntermediateJWT{}, func(t *jwt.Token) (interface{}, error) {
			return signingKey, nil
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

		setSessionCookie(w, sessionCookieRequest{
			oauthCfg:        c.cfg,
			supportsRefresh: false,
			email:           data.Email,
			accessToken:     tokens.AccessToken,
			provider:        c.googleCfg.Name,
			refreshInterval: refreshInterval,
			signingKey:      signingKey,
			encryptionKey:   encryptionKey,
			targetHost:      token.Claims.(*IntermediateJWT).TargetHost,
		})

		http.Redirect(w, r, fmt.Sprintf("%s://%s", scheme, token.Claims.(*IntermediateJWT).TargetHost), http.StatusFound)
	}
	http.Handle(fmt.Sprintf("/%v/auth/callback", c.googleCfg.Name), rp.CodeExchangeHandler(login, provider))

	logout := func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(c.cfg.CookieName)
		if err == nil {
			tkn, err := jwt.ParseWithClaims(cookie.Value, &zrokClaims{}, func(t *jwt.Token) (interface{}, error) {
				return signingKey, nil
			})
			if err == nil {
				claims := tkn.Claims.(*zrokClaims)
				if claims.Provider == c.googleCfg.Name {
					accessToken, err := decryptToken(claims.AccessToken, encryptionKey)
					if err == nil {
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
					logrus.Errorf("expected provider name '%v' got '%v'", c.googleCfg.Name, claims.Provider)
					proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(errors.New("provider name mismatch")))
					return
				}
			} else {
				logrus.Errorf("invalid jwt; unable to parse: %v", err)
				proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("invalid jwt; unable to parse")))
				return
			}
		} else {
			logrus.Errorf("error getting cookie '%v': %v", c.cfg.CookieName, err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("error getting cookie")))
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     c.cfg.CookieName,
			Value:    "",
			MaxAge:   -1,
			Domain:   c.cfg.CookieDomain,
			Path:     "/",
			HttpOnly: true,
		})

		redirectURL := r.URL.Query().Get("redirect_url")
		if redirectURL == "" {
			redirectURL = fmt.Sprintf("%s/%s/login", c.cfg.EndpointUrl, c.googleCfg.Name)
		}
		http.Redirect(w, r, redirectURL, http.StatusFound)
	}
	http.HandleFunc(fmt.Sprintf("/%v/logout", c.googleCfg.Name), logout)

	logrus.Infof("configured google provider at '/%v'", c.googleCfg.Name)

	return nil
}
