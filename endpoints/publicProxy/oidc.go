package publicProxy

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	zhttp "github.com/zitadel/oidc/v3/pkg/http"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"golang.org/x/oauth2"
)

type OIDCProvider struct {
	name          string
	config        *oauth2.Config
	relyingParty  rp.RelyingParty
	emailEndpoint string
	emailPath     string
}

type IntermediateJWT struct {
	State                      string `json:"state"`
	Host                       string `json:"host"`
	AuthorizationCheckInterval string `json:"authorizationCheckInterval"`
	jwt.RegisteredClaims
}

func configureOIDCProvider(cfg *OauthConfig, providerCfg *OauthProviderConfig, tls bool) (*OIDCProvider, error) {
	logrus.Infof("configuring oidc provider: %v", providerCfg.Name)

	if providerCfg == nil {
		return nil, errors.New("provider configuration is required")
	}

	rpConfig := &oauth2.Config{
		ClientID:     providerCfg.ClientId,
		ClientSecret: providerCfg.ClientSecret,
		RedirectURL:  fmt.Sprintf("%v/oauth/%s", cfg.RedirectUrl, providerCfg.Name),
		Scopes:       providerCfg.Scopes,
		Endpoint:     providerCfg.GetEndpoint(),
	}

	hash := md5.New()
	if n, err := hash.Write([]byte(cfg.HashKey)); err != nil {
		return nil, err
	} else if n != len(cfg.HashKey) {
		return nil, errors.New("short hash")
	}
	key := hash.Sum(nil)

	cookieHandler := zhttp.NewCookieHandler(key, key, zhttp.WithUnsecure(), zhttp.WithDomain(cfg.CookieDomain))

	options := []rp.Option{
		rp.WithCookieHandler(cookieHandler),
		rp.WithVerifierOpts(rp.WithIssuedAtOffset(5 * time.Second)),
	}

	if providerCfg.SupportsPKCE {
		options = append(options, rp.WithPKCE(cookieHandler))
	}

	relyingParty, err := rp.NewRelyingPartyOAuth(rpConfig, options...)
	if err != nil {
		return nil, err
	}

	return &OIDCProvider{
		name:          providerCfg.Name,
		config:        rpConfig,
		relyingParty:  relyingParty,
		emailEndpoint: providerCfg.EmailEndpoint,
		emailPath:     providerCfg.EmailPath,
	}, nil
}

func (p *OIDCProvider) setupHandlers(cfg *OauthConfig, key []byte, tls bool) {
	scheme := "http"
	if tls {
		scheme = "https"
	}

	authHandlerWithQueryState := func(party rp.RelyingParty) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			host, err := url.QueryUnescape(r.URL.Query().Get("targethost"))
			if err != nil {
				logrus.Errorf("unable to unescape target host: %v", err)
				deleteZrokCookies(w, r)
				http.Error(w, "Invalid target host", http.StatusBadRequest)
				return
			}

			host = strings.TrimSpace(host)
			if host == "" {
				logrus.Error("target host is empty")
				deleteZrokCookies(w, r)
				http.Error(w, "Empty target host", http.StatusBadRequest)
				return
			}

			if strings.Contains(host, "://") {
				if parsedURL, err := url.Parse(host); err == nil && parsedURL.Host != "" {
					host = parsedURL.Host
				}
			}
			host = strings.Split(host, "/")[0]

			if host == "" {
				logrus.Error("failed to extract valid host")
				deleteZrokCookies(w, r)
				http.Error(w, "Invalid target host", http.StatusBadRequest)
				return
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
				s, err := t.SignedString(key)
				if err != nil {
					logrus.Errorf("unable to sign intermediate JWT: %v", err)
				}
				return s
			}, party, rp.WithURLParam("access_type", "offline"))(w, r)
		}
	}

	getEmail := func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens[*oidc.IDTokenClaims], state string, rp rp.RelyingParty) {
		parsedUrl, err := url.Parse(p.emailEndpoint)
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
			logrus.Errorf("error getting user info from %s: %v", p.name, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() { _ = resp.Body.Close() }()

		response, err := io.ReadAll(resp.Body)
		if err != nil {
			logrus.Errorf("error reading response body: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		email, err := p.extractEmail(response)
		if err != nil {
			logrus.Errorf("error extracting email: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		token, err := jwt.ParseWithClaims(state, &IntermediateJWT{}, func(t *jwt.Token) (interface{}, error) {
			return key, nil
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

		targetHost := token.Claims.(*IntermediateJWT).Host
		logrus.Infof("setting cookie and redirecting to host: %s", targetHost)

		setZrokCookie(w, cfg.CookieDomain, email, tokens.AccessToken, p.name, authCheckInterval, key, targetHost)
		http.Redirect(w, r, fmt.Sprintf("%s://%s", scheme, targetHost), http.StatusFound)
	}

	http.Handle(fmt.Sprintf("/oauth/%s/login", p.name), authHandlerWithQueryState(p.relyingParty))
	http.Handle(fmt.Sprintf("/oauth/%s", p.name), rp.CodeExchangeHandler(getEmail, p.relyingParty))
}

func (p *OIDCProvider) extractEmail(response []byte) (string, error) {
	var data interface{}
	if err := json.Unmarshal(response, &data); err != nil {
		return "", err
	}

	// handle array response (like GitHub's email endpoint)
	if arr, ok := data.([]interface{}); ok {
		for _, item := range arr {
			if email, found := p.findEmailInMap(item.(map[string]interface{})); found {
				return email, nil
			}
		}
		return "", errors.New("no primary email found in array response")
	}

	// handle single object response (like Google's userinfo endpoint)
	if obj, ok := data.(map[string]interface{}); ok {
		if email, found := p.findEmailInMap(obj); found {
			return email, nil
		}
		return "", errors.New("no email found in object response")
	}

	return "", errors.New("unexpected response format")
}

func (p *OIDCProvider) findEmailInMap(obj map[string]interface{}) (string, bool) {
	paths := strings.Split(p.emailPath, ".")
	current := obj

	for i, path := range paths {
		if i == len(paths)-1 {
			if email, ok := current[path].(string); ok {
				return email, true
			}
			return "", false
		}

		if next, ok := current[path].(map[string]interface{}); ok {
			current = next
		} else {
			return "", false
		}
	}

	return "", false
}
