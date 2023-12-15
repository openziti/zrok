package publicProxy

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/zitadel/oidc/v2/pkg/client/rp"
	zhttp "github.com/zitadel/oidc/v2/pkg/http"
	"github.com/zitadel/oidc/v2/pkg/oidc"
	"golang.org/x/oauth2"
	githubOAuth "golang.org/x/oauth2/github"
	"io"
	"net/http"
	"net/url"
	"time"
)

func configureGithubOauth(cfg *OauthConfig, tls bool) error {
	scheme := "http"
	if tls {
		scheme = "https"
	}

	providerCfg := cfg.GetProvider("github")
	if providerCfg == nil {
		logrus.Info("unable to find provider config for github; skipping")
		return nil
	}
	clientID := providerCfg.ClientId
	rpConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: providerCfg.ClientSecret,
		RedirectURL:  fmt.Sprintf("%v/github/oauth", cfg.RedirectUrl),
		Scopes:       []string{"user:email"},
		Endpoint:     githubOAuth.Endpoint,
	}

	hash := md5.New()
	n, err := hash.Write([]byte(cfg.HashKey))
	if err != nil {
		return err
	}
	if n != len(cfg.HashKey) {
		return errors.New("short hash")
	}
	key := hash.Sum(nil)

	cookieHandler := zhttp.NewCookieHandler(key, key, zhttp.WithUnsecure(), zhttp.WithDomain(cfg.CookieDomain))

	options := []rp.Option{
		rp.WithCookieHandler(cookieHandler),
		rp.WithVerifierOpts(rp.WithIssuedAtOffset(5 * time.Second)),
		//rp.WithPKCE(cookieHandler), //Github currently doesn't support pkce. Update when that changes.
	}

	relyingParty, err := rp.NewRelyingPartyOAuth(rpConfig, options...)
	if err != nil {
		return err
	}

	type IntermediateJWT struct {
		State                      string `json:"state"`
		Host                       string `json:"host"`
		AuthorizationCheckInterval string `json:"authorizationCheckInterval"`
		jwt.RegisteredClaims
	}

	type githubUserResp struct {
		Email      string
		Primary    bool
		Verified   bool
		Visibility string
	}

	authHandlerWithQueryState := func(party rp.RelyingParty) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			host, err := url.QueryUnescape(r.URL.Query().Get("targethost"))
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
				s, err := t.SignedString(key)
				if err != nil {
					logrus.Errorf("unable to sign intermediate JWT: %v", err)
				}
				return s
			}, party, rp.WithURLParam("access_type", "offline"))(w, r)
		}
	}

	http.Handle("/github/login", authHandlerWithQueryState(relyingParty))
	getEmail := func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens[*oidc.IDTokenClaims], state string, rp rp.RelyingParty) {
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
			logrus.Errorf("error unmarshalling google oauth response: %v", err)
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
		SetZrokCookie(w, cfg.CookieDomain, primaryEmail, tokens.AccessToken, "github", authCheckInterval, key, token.Claims.(*IntermediateJWT).Host)
		http.Redirect(w, r, fmt.Sprintf("%s://%s", scheme, token.Claims.(*IntermediateJWT).Host), http.StatusFound)
	}

	http.Handle("/github/oauth", rp.CodeExchangeHandler(getEmail, relyingParty))
	return nil
}
