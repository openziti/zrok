package publicProxy

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

func setSessionCookie(w http.ResponseWriter, supportsRefresh bool, cookieDomain, email, accessToken, provider string, checkInterval time.Duration, key []byte, targetHost string) {
	targetHost = strings.TrimSpace(targetHost)
	if targetHost == "" {
		logrus.Error("host claim must not be empty")
		http.Error(w, "host claim must not be empty", http.StatusBadRequest)
		return
	}
	targetHost = strings.Split(targetHost, "/")[0]
	logrus.Debugf("setting zrok-access cookie JWT audience '%s'", targetHost)

	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, &zrokClaims{
		Email:                      email,
		AccessToken:                accessToken,
		SupportsRefresh:            supportsRefresh,
		Provider:                   provider,
		TargetHost:                 targetHost,
		AuthorizationCheckInterval: checkInterval,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(checkInterval).Add(30 * time.Second)),
		},
	})
	sTkn, err := tkn.SignedString(key)
	if err != nil {
		http.Error(w, fmt.Sprintf("after signing cookie token: %v", err.Error()), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "zrok-access",
		Value:   sTkn,
		MaxAge:  int(30 + checkInterval.Seconds()),
		Domain:  cookieDomain,
		Path:    "/",
		Expires: time.Now().Add(checkInterval).Add(30 * time.Second),
		// Secure:  true, // pending server tls feature https://github.com/openziti/zrok/issues/24
		HttpOnly: true,                 // enabled because zrok frontend is the only intended consumer of this cookie, not client-side scripts
		SameSite: http.SameSiteLaxMode, // explicitly set to the default Lax mode which allows the zrok share to be navigated to from another site and receive the cookie
	})
}

func filterSessionCookies(w http.ResponseWriter, r *http.Request) {
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
		if cookie.Name != "zrok-access" && cookie.Name != "pkce" {
			r.AddCookie(cookie)
		}
	}
}
