package publicProxy

import (
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/endpoints"
	"github.com/openziti/zrok/v2/endpoints/proxyUi"
	"github.com/pkg/errors"
)

type sessionCookieRequest struct {
	oauthCfg        *OauthConfig
	supportsRefresh bool
	email           string
	accessToken     string
	provider        string
	refreshInterval time.Duration
	signingKey      []byte
	encryptionKey   []byte
	targetHost      string
}

// getSessionCookie retrieves and reassembles a session cookie using the shared endpoints package
func getSessionCookie(r *http.Request, cookieName string) (*http.Cookie, error) {
	return endpoints.GetSessionCookie(r, cookieName)
}

func setSessionCookie(w http.ResponseWriter, req sessionCookieRequest) {
	targetHost := strings.TrimSpace(req.targetHost)
	if targetHost == "" {
		err := errors.New("targetHost claim must not be empty")
		dl.Error(err)
		proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(err))
		return
	}
	targetHost = strings.Split(targetHost, "/")[0]

	encryptedAccessToken, err := endpoints.EncryptToken(req.accessToken, req.encryptionKey)
	if err != nil {
		dl.Errorf("failed to encrypt access token: %v", err)
		proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("failed to encrypt access token")))
		return
	}

	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, &zrokClaims{
		Email:           req.email,
		AccessToken:     encryptedAccessToken,
		SupportsRefresh: req.supportsRefresh,
		Provider:        req.provider,
		TargetHost:      targetHost,
		RefreshInterval: req.refreshInterval,
		NextRefresh:     time.Now().Add(req.refreshInterval),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(req.oauthCfg.SessionLifetime)),
		},
	})
	sTkn, err := tkn.SignedString(req.signingKey)
	if err != nil {
		dl.Errorf("error signing jwt: %v", err)
		proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(req.email).WithError(errors.New("error signing jwt")))
		return
	}

	// use the shared endpoints package to set the cookie with compression and striping
	if err := endpoints.SetSessionCookie(w, req.oauthCfg.CookieName, sTkn, req.oauthCfg); err != nil {
		dl.Errorf("failed to set session cookie: %v", err)
		proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(req.email).WithError(errors.New("failed to set session cookie")))
		return
	}
}

// clearSessionCookies clears all session cookies using the shared endpoints package
func clearSessionCookies(w http.ResponseWriter, r *http.Request, cookieName string, cfg *OauthConfig) {
	endpoints.ClearSessionCookies(w, r, cookieName, cfg)
}

// filterSessionCookies strips out the configured session cookie and also any `pkce` cookie
func filterSessionCookies(w http.ResponseWriter, r *http.Request, cfg *Config) {
	cookies := r.Cookies()
	r.Header.Del("Cookie")

	if cfg.Oauth != nil {
		// use the shared endpoints package to filter session cookies
		filtered := endpoints.FilterSessionCookies(cookies, cfg.Oauth.CookieName)
		for _, cookie := range filtered {
			r.AddCookie(cookie)
		}
	} else {
		// no oauth config, just filter pkce
		for _, cookie := range cookies {
			if cookie.Name == "pkce" {
				continue
			}
			r.AddCookie(cookie)
		}
	}
}
