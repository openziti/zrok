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
	googleOauth "golang.org/x/oauth2/google"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
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
	redirectUrl := fmt.Sprintf("%s://%s", scheme, cfg.RedirectHost)
	rpConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: providerCfg.ClientSecret,
		RedirectURL:  fmt.Sprintf("%v:%v%v", redirectUrl, cfg.RedirectPort, callbackPath),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     googleOauth.Endpoint,
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

	u, err := url.Parse(redirectUrl)
	if err != nil {
		logrus.Errorf("unable to parse redirect url: %v", err)
		return err
	}
	parts := strings.Split(u.Hostname(), ".")
	domain := parts[len(parts)-2] + "." + parts[len(parts)-1]

	cookieHandler := zhttp.NewCookieHandler(key, key, zhttp.WithUnsecure(), zhttp.WithDomain(domain))

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
		Host                       string `json:"host"`
		AuthorizationCheckInterval string `json:"authorizationCheckInterval"`
		jwt.RegisteredClaims
	}

	type googleOauthEmailResp struct {
		Email string
	}

	authHandlerWithQueryState := func(party rp.RelyingParty) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			host, err := url.QueryUnescape(r.URL.Query().Get("targethost"))
			if err != nil {
				logrus.Errorf("Unable to unescape target host: %v", err)
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
			logrus.Error("Error getting user info from google: " + err.Error() + "\n")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() {
			_ = resp.Body.Close()
		}()
		response, err := io.ReadAll(resp.Body)
		if err != nil {
			logrus.Errorf("Error reading response body: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logrus.Infof("Response from google userinfo endpoint: %s", string(response))
		rDat := googleOauthEmailResp{}
		err = json.Unmarshal(response, &rDat)
		if err != nil {
			logrus.Errorf("Error unmarshalling google oauth response: %v", err)
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

		SetZrokCookie(w, domain, rDat.Email, tokens.AccessToken, "google", authCheckInterval, key)
		http.Redirect(w, r, fmt.Sprintf("%s://%s", scheme, token.Claims.(*IntermediateJWT).Host), http.StatusFound)
	}

	http.Handle(callbackPath, rp.CodeExchangeHandler(getEmail, relyingParty))
	return nil
}
