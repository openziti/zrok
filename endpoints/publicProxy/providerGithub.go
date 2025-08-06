package publicProxy

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
	"github.com/mitchellh/mapstructure"
	"github.com/openziti/zrok/endpoints/proxyUi"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	zhttp "github.com/zitadel/oidc/v3/pkg/http"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"golang.org/x/oauth2"
	githubOAuth "golang.org/x/oauth2/github"
)

type githubConfigurer struct {
	cfg       *OauthConfig
	githubCfg *githubConfig
	tls       bool
}

func newGithubConfigurer(cfg *OauthConfig, tls bool, v map[string]interface{}) (*githubConfigurer, error) {
	c := &githubConfigurer{cfg: cfg}
	githubCfg, err := newGithubConfig(v)
	if err != nil {
		return nil, err
	}
	c.githubCfg = githubCfg
	c.tls = tls
	return c, nil
}

type githubConfig struct {
	Name         string `mapstructure:"name"`
	ClientId     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
}

func newGithubConfig(v map[string]interface{}) (*githubConfig, error) {
	cfg := &githubConfig{}
	if err := mapstructure.Decode(v, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *githubConfigurer) configure() error {
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
		ClientID:     c.githubCfg.ClientId,
		ClientSecret: c.githubCfg.ClientSecret,
		RedirectURL:  fmt.Sprintf("%v/%v/auth/callback", c.cfg.EndpointUrl, c.githubCfg.Name),
		Scopes:       []string{"user:email"},
		Endpoint:     githubOAuth.Endpoint,
	}
	providerOptions := []rp.Option{
		rp.WithCookieHandler(cookieHandler),
		rp.WithVerifierOpts(rp.WithIssuedAtOffset(5 * time.Second)),
	}
	provider, err := rp.NewRelyingPartyOAuth(rpConfig, providerOptions...)
	if err != nil {
		return err
	}

	type githubUserResp struct {
		Email      string
		Primary    bool
		Verified   bool
		Visibility string
	}

	auth := func(provider rp.RelyingParty) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			targetHost, err := url.QueryUnescape(r.URL.Query().Get("targetHost"))
			if err != nil {
				err := fmt.Errorf("unable to unescape targetHost: %v", err)
				logrus.Error(err)
				proxyUi.WriteUnauthorized(w, proxyUi.RequiredData("unauthorized!", "unauthorized!").WithError(err))
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
	http.Handle(fmt.Sprintf("/%v/login", c.githubCfg.Name), auth(provider))

	login := func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens[*oidc.IDTokenClaims], state string, rp rp.RelyingParty) {
		token, err := jwt.ParseWithClaims(state, &IntermediateJWT{}, func(t *jwt.Token) (interface{}, error) {
			return signingKey, nil
		})
		if err != nil {
			errOut := errors.Wrap(err, "error parsing intermediate token")
			logrus.Error(errOut)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errOut))
			return
		}

		var refreshInterval time.Duration
		if v, err := time.ParseDuration(token.Claims.(*IntermediateJWT).RefreshInterval); err == nil {
			refreshInterval = v
		} else {
			errOut := errors.Wrapf(err, "unable to parse authorization check interval '%v'", token.Claims.(*IntermediateJWT).RefreshInterval)
			logrus.Error(errOut)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errOut))
			return
		}

		parsedUrl, err := url.Parse("https://api.github.com/user/emails")
		if err != nil {
			errOut := errors.Wrap(err, "error parsing github url")
			logrus.Error(errOut)
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
			logrus.Error(errOut)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errOut))
			return
		}
		defer func() {
			_ = resp.Body.Close()
		}()
		response, err := io.ReadAll(resp.Body)
		if err != nil {
			errOut := errors.Wrap(err, "error reading response body from github")
			logrus.Error(errOut)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errOut))
			return
		}
		var rDat []githubUserResp
		err = json.Unmarshal(response, &rDat)
		if err != nil {
			errOut := errors.Wrap(err, "error unmarshalling response from github")
			logrus.Error(errOut)
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

		setSessionCookie(w, sessionCookieRequest{
			oauthCfg:        c.cfg,
			supportsRefresh: false,
			email:           primaryEmail,
			accessToken:     tokens.AccessToken,
			provider:        c.githubCfg.Name,
			refreshInterval: refreshInterval,
			signingKey:      signingKey,
			encryptionKey:   encryptionKey,
			targetHost:      token.Claims.(*IntermediateJWT).TargetHost,
		})

		http.Redirect(w, r, fmt.Sprintf("%s://%s", scheme, token.Claims.(*IntermediateJWT).TargetHost), http.StatusFound)
	}
	http.Handle(fmt.Sprintf("/%v/auth/callback", c.githubCfg.Name), rp.CodeExchangeHandler(login, provider))

	logout := func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(c.cfg.CookieName)
		if err == nil {
			tkn, err := jwt.ParseWithClaims(cookie.Value, &zrokClaims{}, func(t *jwt.Token) (interface{}, error) {
				return signingKey, nil
			})
			if err == nil {
				claims := tkn.Claims.(*zrokClaims)
				if claims.Provider == c.githubCfg.Name {
					accessToken, err := decryptToken(claims.AccessToken, encryptionKey)
					if err == nil {
						req, err := http.NewRequest("DELETE",
							fmt.Sprintf("https://api.github.com/applications/%s/token", c.githubCfg.ClientId),
							strings.NewReader(fmt.Sprintf(`{"access_token":"%s"}`, accessToken)))
						if err != nil {
							logrus.Errorf("error creating access token delete request for '%v': %v", claims.Email, err)
							proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(errors.New("error creating access token delete request")))
							return
						}

						req.Header.Set("Content-Type", "application/json")
						req.SetBasicAuth(c.githubCfg.ClientId, c.githubCfg.ClientSecret) // Need client credentials

						resp, err := http.DefaultClient.Do(req)
						if err != nil {
							logrus.Errorf("error invoking access token delete request for '%v': %v", claims.Email, err)
							proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(errors.New("error executing access token delete request")))
							return
						}
						defer resp.Body.Close()

						if resp.StatusCode == http.StatusNoContent {
							logrus.Infof("revoked github access token for '%v'", claims.Email)
						} else {
							logrus.Errorf("access token revocation failed with status: %v", resp.StatusCode)
							proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(errors.New("access token revocation failed")))
							return
						}
					} else {
						logrus.Errorf("unable to decrypt access token for '%v': %v", claims.Email, err)
						proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(errors.New("unable to decrypt access token")))
						return
					}
				} else {
					logrus.Errorf("expected provider name '%v' got '%v'", c.githubCfg.Name, claims.Email)
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
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("invalid cookie")))
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
			redirectURL = fmt.Sprintf("%s/%s/login", c.cfg.EndpointUrl, c.githubCfg.Name)
		}
		http.Redirect(w, r, redirectURL, http.StatusFound)
	}
	http.HandleFunc(fmt.Sprintf("/%v/logout", c.githubCfg.Name), logout)

	logrus.Infof("configured github provider at '/%v", c.githubCfg.Name)

	return nil
}
