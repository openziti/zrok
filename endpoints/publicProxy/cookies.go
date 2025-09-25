package publicProxy

import (
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/endpoints/proxyUi"
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

func setSessionCookie(w http.ResponseWriter, req sessionCookieRequest) {
	targetHost := strings.TrimSpace(req.targetHost)
	if targetHost == "" {
		err := errors.New("targetHost claim must not be empty")
		dl.Error(err)
		proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(err))
		return
	}
	targetHost = strings.Split(targetHost, "/")[0]

	encryptedAccessToken, err := encryptToken(req.accessToken, req.encryptionKey)
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

	http.SetCookie(w, &http.Cookie{
		Name:    req.oauthCfg.CookieName,
		Value:   sTkn,
		MaxAge:  int(req.oauthCfg.SessionLifetime.Seconds()),
		Domain:  req.oauthCfg.CookieDomain,
		Path:    "/",
		Expires: time.Now().Add(req.oauthCfg.SessionLifetime),
		// Secure:  true, // pending server tls feature https://github.com/openziti/zrok/issues/24
		HttpOnly: true,                 // enabled because zrok frontend is the only intended consumer of this cookie, not client-side scripts
		SameSite: http.SameSiteLaxMode, // explicitly set to the default Lax mode which allows the zrok share to be navigated to from another site and receive the cookie
	})
}

// filterSessionCookies strips out the configured session cookie and also any `pkce` cookie
func filterSessionCookies(w http.ResponseWriter, r *http.Request, cfg *Config) {
	cookies := r.Cookies()
	r.Header.Del("Cookie")
	for _, cookie := range cookies {
		if cfg.Oauth != nil && cfg.Oauth.CookieName == cookie.Name {
			continue
		}
		if cookie.Name == "pkce" {
			continue
		}
		r.AddCookie(cookie)
	}
}
