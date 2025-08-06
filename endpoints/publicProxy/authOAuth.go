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
	Email           string        `json:"em"`
	AccessToken     string        `json:"acc"`
	SupportsRefresh bool          `json:"srf"`
	Provider        string        `json:"pr"`
	TargetHost      string        `json:"th"`
	RefreshInterval time.Duration `json:"rfi"`
	NextRefresh     time.Time     `json:"nr"`
	jwt.RegisteredClaims
}

func (c *zrokClaims) getTargetHost() (jwt.ClaimStrings, error) {
	return jwt.ClaimStrings{c.TargetHost}, nil
}

func oauthLoginRequired(w http.ResponseWriter, r *http.Request, cfg *OauthConfig, provider, target string, refreshInterval time.Duration) {
	http.Redirect(w, r, fmt.Sprintf("%s/%s/login?targetHost=%s&refreshInterval=%s", cfg.EndpointUrl, provider, url.QueryEscape(target), refreshInterval.String()), http.StatusFound)
}

func oauthRefreshRequired(w http.ResponseWriter, r *http.Request, cfg *OauthConfig, provider, target string) {
	http.Redirect(w, r, fmt.Sprintf("%s/%s/refresh?targetHost=%s", cfg.EndpointUrl, provider, url.QueryEscape(target)), http.StatusFound)
}

func (h *authHandler) handleOAuth(w http.ResponseWriter, r *http.Request, cfg map[string]interface{}, shrToken string) bool {
	oauthCfg, found := cfg["oauth"]
	if !found {
		logrus.Warnf("%v -> no oauth cfg for '%v'", r.RemoteAddr, shrToken)
		return false
	}

	oauthMap := oauthCfg.(map[string]interface{})
	provider := oauthMap["provider"].(string)
	refreshInterval := getRefreshInterval(oauthMap)
	target := fmt.Sprintf("%s%s", r.Host, r.URL.Path)

	cookie, err := r.Cookie(h.cfg.Oauth.CookieName)
	if err != nil {
		logrus.Errorf("unable to get '%v' cookie: %v", h.cfg.Oauth.CookieName, err)
		oauthLoginRequired(w, r, h.cfg.Oauth, provider, target, refreshInterval)
		return false
	}

	if !h.validateOAuthToken(w, r, cookie, provider, refreshInterval, target) {
		return false
	}

	if !h.validateEmailDomain(w, oauthMap, cookie, h.cfg) {
		return false
	}

	return true
}

func (h *authHandler) validateOAuthToken(w http.ResponseWriter, r *http.Request, cookie *http.Cookie, provider string, refreshInterval time.Duration, target string) bool {
	tkn, err := jwt.ParseWithClaims(cookie.Value, &zrokClaims{}, func(t *jwt.Token) (interface{}, error) {
		if h.cfg.Oauth == nil {
			return nil, fmt.Errorf("missing oauth configuration for access point; unable to parse jwt")
		}
		return h.signingKey, nil
	})
	if err != nil {
		logrus.Errorf("unable to parse jwt: %v", err)
		oauthLoginRequired(w, r, h.cfg.Oauth, provider, target, refreshInterval)
		return false
	}

	claims := tkn.Claims.(*zrokClaims)
	if claims.Provider != provider || claims.RefreshInterval != refreshInterval || claims.TargetHost != r.Host {
		logrus.Errorf("token validation failed; restarting auth flow (email: '%v', target: '%v')", claims.Email, target)
		oauthLoginRequired(w, r, h.cfg.Oauth, provider, target, refreshInterval)
		return false
	}

	if time.Now().After(claims.NextRefresh) {
		if claims.SupportsRefresh {
			logrus.Infof("oauth session expired; refreshing tokens (email: '%v', target: '%v')", claims.Email, target)
			oauthRefreshRequired(w, r, h.cfg.Oauth, provider, target)
		} else {
			logrus.Warnf("oauth session expired; re-login (email: '%v', target: '%v')", claims.Email, target)
			oauthLoginRequired(w, r, h.cfg.Oauth, provider, target, refreshInterval)
		}
		return false
	} else {
		logrus.Debugf("%v until next refresh", time.Until(claims.NextRefresh))
	}

	r.Header.Set("zrok-auth-provider", provider)
	r.Header.Set("zrok-auth-email", claims.Email)
	r.Header.Set("zrok-auth-expires", claims.NextRefresh.Format(time.RFC3339))

	return true
}

func (h *authHandler) validateEmailDomain(w http.ResponseWriter, oauthCfg map[string]interface{}, cookie *http.Cookie, cfg *Config) bool {
	if patterns, found := oauthCfg["email_domains"].([]interface{}); found && len(patterns) > 0 {
		tkn, _ := jwt.ParseWithClaims(cookie.Value, &zrokClaims{}, func(t *jwt.Token) (interface{}, error) {
			return h.signingKey, nil
		})
		claims := tkn.Claims.(*zrokClaims)

		for _, pattern := range patterns {
			if castedPattern, ok := pattern.(string); ok {
				match, err := glob.Compile(castedPattern)
				if err != nil {
					err := fmt.Errorf("invalid email address pattern glob '%v': %v", pattern, err)
					logrus.Error(err)
					proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(err))
					return false
				}
				if match.Match(claims.Email) {
					return true
				}
			}
		}
		logrus.Warnf("unauthorized email '%v'", claims.Email)
		proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email))
		return false
	}
	return true
}

func getRefreshInterval(oauthCfg map[string]interface{}) time.Duration {
	if refreshInterval, found := oauthCfg["authorization_check_interval"]; !found {
		logrus.Error("missing 'authorization_check_interval', defaulting to 3 hours")
		return 3 * time.Hour
	} else {
		i, err := time.ParseDuration(refreshInterval.(string))
		if err != nil {
			logrus.Errorf("invalid refresh interval '%v', defaulting to 3 hours", refreshInterval)
			return 3 * time.Hour
		}
		return i
	}
}
