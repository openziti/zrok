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
	"github.com/zitadel/oidc/v2/pkg/client/rp"
	zhttp "github.com/zitadel/oidc/v2/pkg/http"
	"github.com/zitadel/oidc/v2/pkg/oidc"
	"golang.org/x/oauth2"
	googleOauth "golang.org/x/oauth2/google"
)

type googleOauthConfigurer struct {
	cfg      *OauthConfig
	oauthCfg *googleOauthConfig
	tls      bool
}

func newGoogleOauthConfigurer(cfg *OauthConfig, tls bool, v map[string]interface{}) (*googleOauthConfigurer, error) {
	c := &googleOauthConfigurer{cfg: cfg}
	oauthCfg, err := newGoogleOauthConfig(v)
	if err != nil {
		return nil, err
	}
	c.oauthCfg = oauthCfg
	c.tls = tls
	return c, nil
}

type googleOauthConfig struct {
	ClientId      string   `mapstructure:"client_id"`
	ClientSecret  string   `mapstructure:"client_secret"`
	Scopes        []string `mapstructure:"scopes"`
	AuthUrl       string   `mapstructure:"auth_url"`
	TokenUrl      string   `mapstructure:"token_url"`
	EmailEndpoint string   `mapstructure:"email_endpoint"`
	EmailPath     string   `mapstructure:"email_path"`
	SupportsPkce  bool     `mapstructure:"supports_pkce"`
}

func newGoogleOauthConfig(v map[string]interface{}) (*googleOauthConfig, error) {
	cfg := &googleOauthConfig{}
	if err := mapstructure.Decode(v, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *googleOauthConfigurer) configure() error {
	scheme := "http"
	if c.tls {
		scheme = "https"
	}

	clientID := c.oauthCfg.ClientId
	rpConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: c.oauthCfg.ClientSecret,
		RedirectURL:  fmt.Sprintf("%v/google/auth/callback", c.cfg.EndpointUrl),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     googleOauth.Endpoint,
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

	auth := func(party rp.RelyingParty) http.HandlerFunc {
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
			}, party, rp.WithURLParam("access_type", "offline"), rp.URLParamOpt(rp.WithPrompt("login")))(w, r)
		}
	}
	http.Handle("/google/login", auth(provider))

	getEmail := func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens[*oidc.IDTokenClaims], state string, rp rp.RelyingParty) {
		resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + url.QueryEscape(tokens.AccessToken))
		if err != nil {
			logrus.Errorf("error getting user info from google: %v", err)
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
		logrus.Infof("response from google userinfo endpoint: %s", string(response))
		rDat := googleOauthEmailResp{}
		err = json.Unmarshal(response, &rDat)
		if err != nil {
			logrus.Errorf("error unmarshalling google oauth response: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		token, err := jwt.ParseWithClaims(state, &IntermediateJWT{}, func(t *jwt.Token) (interface{}, error) {
			return signingKey, nil
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("After intermediate token parse: %v", err.Error()), http.StatusInternalServerError)
			return
		}

		authCheckInterval := 3 * time.Hour
		i, err := time.ParseDuration(token.Claims.(*IntermediateJWT).AuthorizationCheckInterval)
		if err != nil {
			logrus.Errorf("unable to parse authorization check interval: %v. Defaulting to 3 hours", err)
		} else {
			authCheckInterval = i
		}
		setSessionCookie(w, false, c.cfg.CookieDomain, rDat.Email, tokens.AccessToken, "google", authCheckInterval, signingKey, encryptionKey, token.Claims.(*IntermediateJWT).Host)
		http.Redirect(w, r, fmt.Sprintf("%s://%s", scheme, token.Claims.(*IntermediateJWT).Host), http.StatusFound)
	}
	http.Handle("/google/auth/callback", rp.CodeExchangeHandler(getEmail, provider))

	logrus.Info("configured google provider at '/google'")

	return nil
}
