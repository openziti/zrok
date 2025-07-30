package publicProxy

import (
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/openziti/zrok/endpoints/proxyUi"
	"github.com/sirupsen/logrus"
)

func setSessionCookie(w http.ResponseWriter, cfg *OauthConfig, supportsRefresh bool, email, accessToken, provider string, checkInterval time.Duration, signingKey []byte, encryptionKey []byte, targetHost string) {
	targetHost = strings.TrimSpace(targetHost)
	if targetHost == "" {
		logrus.Error("targetHost claim must not be empty")
		proxyUi.WriteUnauthorized(w)
		return
	}
	targetHost = strings.Split(targetHost, "/")[0]
	logrus.Debugf("setting zrok-access cookie JWT audience '%s'", targetHost)

	encryptedAccessToken, err := encryptToken(accessToken, encryptionKey)
	if err != nil {
		logrus.Errorf("failed to encrypt access token: %v", err)
		proxyUi.WriteUnauthorized(w)
		return
	}

	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, &zrokClaims{
		Email:                      email,
		AccessToken:                encryptedAccessToken,
		SupportsRefresh:            supportsRefresh,
		Provider:                   provider,
		TargetHost:                 targetHost,
		AuthorizationCheckInterval: checkInterval,
		NextRefresh:                time.Now().Add(checkInterval),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.SessionLifetime)),
		},
	})
	sTkn, err := tkn.SignedString(signingKey)
	if err != nil {
		logrus.Errorf("error signing jwt: %v", err)
		proxyUi.WriteUnauthorized(w)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    cfg.CookieName,
		Value:   sTkn,
		MaxAge:  int(cfg.SessionLifetime.Seconds()),
		Domain:  cfg.CookieDomain,
		Path:    "/",
		Expires: time.Now().Add(cfg.SessionLifetime),
		// Secure:  true, // pending server tls feature https://github.com/openziti/zrok/issues/24
		HttpOnly: true,                 // enabled because zrok frontend is the only intended consumer of this cookie, not client-side scripts
		SameSite: http.SameSiteLaxMode, // explicitly set to the default Lax mode which allows the zrok share to be navigated to from another site and receive the cookie
	})
}

func filterSessionCookies(w http.ResponseWriter, r *http.Request, cfg *Config) {
	// Get all cookies from the request
	cookies := r.Cookies()
	// Clear the Cookie header
	r.Header.Del("Cookie")
	// Save cookies not in the list of cookies to delete, the pkce cookie might be okay to pass along to the HTTP
	// backend, but zrok-access is not because it can contain the accessToken from any other OAuth enabled shares, so we
	// delete it here when the current share is not OAuth-enabled. OAuth-enabled shares check the audience claim in the
	// JWT to ensure it matches the requested share and will send the client back to the OAuth provider if it does not
	// match.
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
