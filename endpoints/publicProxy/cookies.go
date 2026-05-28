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
func getSessionCookie(r *http.Request, cfg *OauthConfig) (*http.Cookie, error) {
	return endpoints.GetSessionCookie(r, cfg)
}

// buildSessionJWT creates and signs the session JWT from a sessionCookieRequest.
// Returns the signed JWT string or an error.
func buildSessionJWT(req sessionCookieRequest) (string, error) {
	targetHost := strings.TrimSpace(req.targetHost)
	if targetHost == "" {
		return "", errors.New("targetHost claim must not be empty")
	}
	targetHost = strings.Split(targetHost, "/")[0]

	encryptedAccessToken, err := endpoints.EncryptToken(req.accessToken, req.encryptionKey)
	if err != nil {
		return "", errors.Wrap(err, "failed to encrypt access token")
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
	return tkn.SignedString(req.signingKey)
}

// setSessionCookie builds, signs, and sets the session cookie. Returns the signed
// JWT string so callers can inspect it (e.g. for the ReturnToken display path).
// On error the response is written and "" is returned.
func setSessionCookie(w http.ResponseWriter, req sessionCookieRequest) string {
	sTkn, err := buildSessionJWT(req)
	if err != nil {
		dl.Errorf("error building session jwt: %v", err)
		proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(errors.New("error building session jwt")))
		return ""
	}

	// use the shared endpoints package to set the cookie with compression and striping
	if err := endpoints.SetSessionCookie(w, req.oauthCfg.CookieName, sTkn, req.oauthCfg); err != nil {
		dl.Errorf("failed to set session cookie: %v", err)
		proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(req.email).WithError(errors.New("failed to set session cookie")))
		return ""
	}

	return sTkn
}

// clearSessionCookies clears all session cookies using the shared endpoints package
func clearSessionCookies(w http.ResponseWriter, r *http.Request, cookieName string, cfg *OauthConfig) {
	endpoints.ClearSessionCookies(w, r, cookieName, cfg)
}

// filterSessionCookies strips out the configured session cookie, any pkce cookie,
// and the X-Zrok-Session header from the request before proxying to the backend.
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

	endpoints.StripSessionHeader(r)
}
