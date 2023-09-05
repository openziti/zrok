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
	googleOauth "golang.org/x/oauth2/google"
)

func configureGoogleOauth(cfg *OauthConfig, tls bool) error {
	scheme := "http"
	if tls {
		scheme = "https"
	}

	providerCfg := cfg.GetProvider("google")
	if providerCfg == nil {
		logrus.Info("unable to find provider config for google. Skipping.")
		return nil
	}

	clientID := providerCfg.ClientId
	callbackPath := "/google/oauth"
	port := cfg.Port
	redirectUrl := fmt.Sprintf("%s://%s", scheme, cfg.RedirectUrl)
	rpConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: providerCfg.ClientSecret,
		RedirectURL:  fmt.Sprintf("%v:%v%v", redirectUrl, port, callbackPath),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     googleOauth.Endpoint,
	}

	key := []byte(cfg.HashKeyRaw)

	cookieHandler := zhttp.NewCookieHandler(key, key, zhttp.WithUnsecure(), zhttp.WithDomain(cfg.RedirectUrl))

	options := []rp.Option{
		rp.WithCookieHandler(cookieHandler),
		rp.WithVerifierOpts(rp.WithIssuedAtOffset(5 * time.Second)),
		rp.WithPKCE(cookieHandler),
	}

	relyingParty, err := rp.NewRelyingPartyOAuth(rpConfig, options...)
	if err != nil {
		return err
	}

	type IntermediateJWT struct {
		State                      string `json:"state"`
		Share                      string `json:"share"`
		AuthorizationCheckInterval string `json:"authorizationCheckInterval"`
		jwt.RegisteredClaims
	}

	type googleOauthEmailResp struct {
		Email string
	}

	authHandlerWithQueryState := func(party rp.RelyingParty) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			rp.AuthURLHandler(func() string {
				id := uuid.New().String()
				t := jwt.NewWithClaims(jwt.SigningMethodHS256, IntermediateJWT{
					id,
					r.URL.Query().Get("share"),
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
					logrus.Errorf("Unable to sign intermediate JWT: %v", err)
				}
				return s
			}, party, rp.WithURLParam("access_type", "offline"))(w, r)
		}
	}

	http.Handle("/google/login", authHandlerWithQueryState(relyingParty))
	getEmail := func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens[*oidc.IDTokenClaims], state string, rp rp.RelyingParty) {
		resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + url.QueryEscape(tokens.AccessToken))
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
		rDat := googleOauthEmailResp{}
		err = json.Unmarshal(response, &rDat)
		if err != nil {
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

		SetZrokCookie(w, rDat.Email, tokens.AccessToken, "google", authCheckInterval, key)
		http.Redirect(w, r, fmt.Sprintf("%s://%s.%s:8080", scheme, token.Claims.(*IntermediateJWT).Share, cfg.RedirectUrl), http.StatusFound)
	}

	http.Handle(callbackPath, rp.CodeExchangeHandler(getEmail, relyingParty))
	return nil
}
