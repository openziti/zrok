package publicProxy

import (
	"encoding/json"
	"fmt"
	"io"
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
		RedirectURL:  fmt.Sprintf("%v/github/auth/callback", c.cfg.EndpointUrl),
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
			host, err := url.QueryUnescape(r.URL.Query().Get("targetHost"))
			if err != nil {
				logrus.Errorf("unable to unescape target host: %v", err)
			}
			rp.AuthURLHandler(func() string {
				id := uuid.New().String()
				t := jwt.NewWithClaims(jwt.SigningMethodHS256, IntermediateJWT{
					id,
					host,
					r.URL.Query().Get("refreshInterval"),
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
			}, provider, rp.WithURLParam("access_type", "offline"), rp.URLParamOpt(rp.WithPrompt("login")))(w, r)
		}
	}
	http.Handle("/github/login", auth(provider))

	login := func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens[*oidc.IDTokenClaims], state string, rp rp.RelyingParty) {
		parsedUrl, err := url.Parse("https://api.github.com/user/emails")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
			logrus.Errorf("error getting user info from github: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() {
			_ = resp.Body.Close()
		}()
		response, err := io.ReadAll(resp.Body)
		if err != nil {
			logrus.Errorf("error reading response body: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var rDat []githubUserResp
		err = json.Unmarshal(response, &rDat)
		if err != nil {
			logrus.Errorf("error unmarshalling github oauth response: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		primaryEmail := ""
		for _, email := range rDat {
			if email.Primary {
				primaryEmail = email.Email
				break
			}
		}

		token, err := jwt.ParseWithClaims(state, &IntermediateJWT{}, func(t *jwt.Token) (interface{}, error) {
			return signingKey, nil
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("after intermediate token parse: %v", err.Error()), http.StatusInternalServerError)
			return
		}

		refreshInterval := 3 * time.Hour
		i, err := time.ParseDuration(token.Claims.(*IntermediateJWT).RefreshInterval)
		if err != nil {
			logrus.Errorf("unable to parse authorization check interval: %v. Defaulting to 3 hours", err)
		} else {
			refreshInterval = i
		}

		setSessionCookie(w, sessionCookieRequest{
			cfg:             c.cfg,
			supportsRefresh: false,
			email:           primaryEmail,
			accessToken:     tokens.AccessToken,
			provider:        "github",
			refreshInterval: refreshInterval,
			signingKey:      signingKey,
			encryptionKey:   encryptionKey,
			targetHost:      token.Claims.(*IntermediateJWT).Host,
		})

		http.Redirect(w, r, fmt.Sprintf("%s://%s", scheme, token.Claims.(*IntermediateJWT).Host), http.StatusFound)
	}
	http.Handle("/github/auth/callback", rp.CodeExchangeHandler(login, provider))

	logrus.Info("configured github provider at '/github")

	return nil
}
