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
	"github.com/sirupsen/logrus"
	"github.com/zitadel/oidc/v2/pkg/client/rp"
	zhttp "github.com/zitadel/oidc/v2/pkg/http"
	"github.com/zitadel/oidc/v2/pkg/oidc"
	"golang.org/x/oauth2"
	githubOAuth "golang.org/x/oauth2/github"
)

func configureGithubOauth(cfg *OauthConfig, tls bool) error {
	scheme := "http"
	if tls {
		scheme = "https"
	}

	providerCfg := cfg.GetProvider("github")
	if providerCfg == nil {
		logrus.Info("unable to find provider config for github. Skipping.")
		return nil
	}
	clientID := providerCfg.ClientId
	callbackPath := "/github/oauth"
	port := cfg.Port
	redirectUrl := fmt.Sprintf("%s://%s", scheme, cfg.RedirectUrl)
	rpConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: providerCfg.ClientSecret,
		RedirectURL:  fmt.Sprintf("%v:%v%v", redirectUrl, port, callbackPath),
		Scopes:       []string{"user:email"},
		Endpoint:     githubOAuth.Endpoint,
	}

	key := []byte(cfg.HashKeyRaw)

	cookieHandler := zhttp.NewCookieHandler(key, key, zhttp.WithUnsecure(), zhttp.WithDomain(cfg.RedirectUrl))

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
		State string `json:"state"`
		Share string `json:"share"`
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
			rp.AuthURLHandler(func() string {
				id := uuid.New().String()
				t := jwt.NewWithClaims(jwt.SigningMethodHS256, IntermediateJWT{
					id,
					r.URL.Query().Get("share"),
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
					logrus.Errorf("Unable to sign intermediate JWT: %v", err)
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
			logrus.Error("Get: " + err.Error() + "\n")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		response, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		rDat := []githubUserResp{}
		err = json.Unmarshal(response, &rDat)
		if err != nil {
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

		SetZrokCookie(w, primaryEmail, tokens.AccessToken, "github", 3*time.Hour, key)

		token, err := jwt.ParseWithClaims(state, &IntermediateJWT{}, func(t *jwt.Token) (interface{}, error) {
			return key, nil
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("After intermediate token parse: %v", err.Error()), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("%s://%s.%s:8080", scheme, token.Claims.(*IntermediateJWT).Share, cfg.RedirectUrl), http.StatusFound)
	}

	http.Handle(callbackPath, rp.CodeExchangeHandler(getEmail, relyingParty))
	return nil
}
