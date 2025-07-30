package publicProxy

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gobwas/glob"
	"github.com/golang-jwt/jwt/v5"
	"github.com/openziti/zrok/endpoints/proxyUi"
	"github.com/sirupsen/logrus"
)

type zrokClaims struct {
	Email                      string        `json:"email"`
	AccessToken                string        `json:"accessToken"`
	Provider                   string        `json:"provider"`
	Audience                   string        `json:"aud"`
	AuthorizationCheckInterval time.Duration `json:"authorizationCheckInterval"`
	jwt.RegisteredClaims
}

func (c *zrokClaims) GetAudience() (jwt.ClaimStrings, error) {
	return jwt.ClaimStrings{c.Audience}, nil
}

func oauthLoginRequired(w http.ResponseWriter, r *http.Request, cfg *OauthConfig, provider, target string, authCheckInterval time.Duration) {
	http.Redirect(w, r, fmt.Sprintf("%s/%s/login?targetHost=%s&checkInterval=%s", cfg.RedirectUrl, provider, url.QueryEscape(target), authCheckInterval.String()), http.StatusFound)
}

func (h *authHandler) handleOAuth(w http.ResponseWriter, r *http.Request, cfg map[string]interface{}, shrToken string) bool {
	oauthCfg, found := cfg["oauth"]
	if !found {
		logrus.Warnf("%v -> no oauth cfg for '%v'", r.RemoteAddr, shrToken)
		return false
	}

	oauthMap := oauthCfg.(map[string]interface{})
	provider := oauthMap["provider"].(string)
	authCheckInterval := getAuthCheckInterval(oauthMap)
	target := fmt.Sprintf("%s%s", r.Host, r.URL.Path)

	cookie, err := r.Cookie("zrok-access")
	if err != nil {
		logrus.Errorf("unable to get 'zrok-access' cookie: %v", err)
		oauthLoginRequired(w, r, h.cfg.Oauth, provider, target, authCheckInterval)
		return false
	}

	if !h.validateOAuthToken(w, r, cookie, provider, authCheckInterval, target) {
		return false
	}

	if !h.validateEmailDomain(w, oauthMap, cookie) {
		return false
	}

	return true
}

func (h *authHandler) validateOAuthToken(w http.ResponseWriter, r *http.Request, cookie *http.Cookie, provider string, authCheckInterval time.Duration, target string) bool {
	tkn, err := jwt.ParseWithClaims(cookie.Value, &zrokClaims{}, func(t *jwt.Token) (interface{}, error) {
		if h.cfg.Oauth == nil {
			return nil, fmt.Errorf("missing oauth configuration for access point; unable to parse jwt")
		}
		return h.key, nil
	})
	if err != nil {
		logrus.Errorf("unable to parse jwt: %v", err)
		oauthLoginRequired(w, r, h.cfg.Oauth, provider, target, authCheckInterval)
		return false
	}

	claims := tkn.Claims.(*zrokClaims)
	if claims.Provider != provider || claims.AuthorizationCheckInterval != authCheckInterval || claims.Audience != r.Host {
		logrus.Error("token validation failed; restarting auth flow")
		oauthLoginRequired(w, r, h.cfg.Oauth, provider, target, authCheckInterval)
		return false
	}

	return true
}

func (h *authHandler) validateEmailDomain(w http.ResponseWriter, oauthCfg map[string]interface{}, cookie *http.Cookie) bool {
	if patterns, found := oauthCfg["email_domains"].([]interface{}); found && len(patterns) > 0 {
		tkn, _ := jwt.ParseWithClaims(cookie.Value, &zrokClaims{}, func(t *jwt.Token) (interface{}, error) {
			return h.key, nil
		})
		claims := tkn.Claims.(*zrokClaims)

		for _, pattern := range patterns {
			if castedPattern, ok := pattern.(string); ok {
				match, err := glob.Compile(castedPattern)
				if err != nil {
					logrus.Errorf("invalid email address pattern glob '%v': %v", pattern, err)
					proxyUi.WriteUnauthorized(w)
					return false
				}
				if match.Match(claims.Email) {
					return true
				}
			}
		}
		logrus.Warnf("unauthorized email '%v'", claims.Email)
		proxyUi.WriteUnauthorized(w)
		return false
	}
	return true
}

func getAuthCheckInterval(oauthCfg map[string]interface{}) time.Duration {
	if checkInterval, found := oauthCfg["authorization_check_interval"]; !found {
		logrus.Error("missing authorization check interval, defaulting to 3 hours")
		return 3 * time.Hour
	} else {
		i, err := time.ParseDuration(checkInterval.(string))
		if err != nil {
			logrus.Errorf("invalid check interval '%v', defaulting to 3 hours", checkInterval)
			return 3 * time.Hour
		}
		return i
	}
}
