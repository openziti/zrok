package publicProxy

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gobwas/glob"
	"github.com/golang-jwt/jwt/v5"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/endpoints"
	"github.com/openziti/zrok/v2/endpoints/proxyUi"
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

func oauthUnauthorized(w http.ResponseWriter, r *http.Request, cfg *OauthConfig, provider, target string, refreshInterval time.Duration) {
	loginURL := fmt.Sprintf("%s/%s/login?targetHost=%s&refreshInterval=%s&return_token=true",
		cfg.EndpointUrl, provider, url.QueryEscape(target), refreshInterval.String())
	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Vary", "Origin")
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", loginURL)
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = fmt.Fprintf(w, `{"login_url":%q}`, loginURL)
}

func (h *authHandler) handleOAuth(w http.ResponseWriter, r *http.Request, cfg map[string]interface{}, shrToken string) bool {
	oauthCfg, found := cfg["oauth"]
	if !found {
		dl.Warnf("%v -> no oauth cfg for '%v'", r.RemoteAddr, shrToken)
		return false
	}

	oauthMap := oauthCfg.(map[string]interface{})
	providerName := oauthMap["provider"].(string)
	refreshInterval := getRefreshInterval(oauthMap)
	target := fmt.Sprintf("%s%s", r.Host, r.URL.Path)

	jwtString := endpoints.GetSessionHeader(r)
	fromHeader := jwtString != ""
	if !fromHeader {
		cookie, err := getSessionCookie(r, h.cfg.Oauth)
		if err != nil {
			dl.Errorf("unable to get '%v' session: %v", h.cfg.Oauth.CookieName, err)
			oauthLoginRequired(w, r, h.cfg.Oauth, providerName, target, refreshInterval)
			return false
		}
		jwtString = cookie.Value
	}

	if !h.validateOAuthToken(w, r, jwtString, fromHeader, providerName, refreshInterval, target) {
		return false
	}

	if !h.validateEmailDomain(w, oauthMap, jwtString) {
		return false
	}

	return true
}

func (h *authHandler) validateOAuthToken(w http.ResponseWriter, r *http.Request, jwtString string, fromHeader bool, provider string, refreshInterval time.Duration, target string) bool {
	tkn, err := jwt.ParseWithClaims(jwtString, &zrokClaims{}, func(t *jwt.Token) (interface{}, error) {
		if h.cfg.Oauth == nil {
			return nil, fmt.Errorf("missing oauth configuration for access point; unable to parse jwt")
		}
		return h.signingKey, nil
	})
	if err != nil {
		dl.Errorf("unable to parse jwt: %v", err)
		if fromHeader {
			oauthUnauthorized(w, r, h.cfg.Oauth, provider, target, refreshInterval)
		} else {
			oauthLoginRequired(w, r, h.cfg.Oauth, provider, target, refreshInterval)
		}
		return false
	}

	claims := tkn.Claims.(*zrokClaims)
	if claims.Provider != provider || claims.RefreshInterval != refreshInterval || claims.TargetHost != r.Host {
		dl.Errorf("token validation failed; restarting auth flow (email: '%v', target: '%v')", claims.Email, target)
		if fromHeader {
			oauthUnauthorized(w, r, h.cfg.Oauth, provider, target, refreshInterval)
		} else {
			oauthLoginRequired(w, r, h.cfg.Oauth, provider, target, refreshInterval)
		}
		return false
	}

	if time.Now().After(claims.NextRefresh) {
		if fromHeader {
			if claims.SupportsRefresh {
				if newJWT, err := h.tryInlineRefresh(claims); err == nil {
					dl.Infof("inline oidc token refresh succeeded for '%v'", claims.Email)
					w.Header().Set(endpoints.SessionHeaderName, newJWT)
					// fall through — request is still authorized with the original claims
				} else {
					dl.Warnf("inline oidc token refresh failed for '%v': %v", claims.Email, err)
					oauthUnauthorized(w, r, h.cfg.Oauth, provider, target, refreshInterval)
					return false
				}
			} else {
				dl.Warnf("oauth session expired; re-authentication required (email: '%v', target: '%v')", claims.Email, target)
				oauthUnauthorized(w, r, h.cfg.Oauth, provider, target, refreshInterval)
				return false
			}
		} else {
			if claims.SupportsRefresh {
				dl.Infof("oauth session expired; refreshing tokens (email: '%v', target: '%v')", claims.Email, target)
				oauthRefreshRequired(w, r, h.cfg.Oauth, provider, target)
			} else {
				dl.Warnf("oauth session expired; re-login (email: '%v', target: '%v')", claims.Email, target)
				oauthLoginRequired(w, r, h.cfg.Oauth, provider, target, refreshInterval)
			}
			return false
		}
	} else {
		dl.Debugf("%v until next refresh", time.Until(claims.NextRefresh))
	}

	r.Header.Set("zrok-auth-provider", provider)
	r.Header.Set("zrok-auth-email", claims.Email)
	r.Header.Set("zrok-auth-expires", claims.NextRefresh.Format(time.RFC3339))

	return true
}

func (h *authHandler) tryInlineRefresh(claims *zrokClaims) (string, error) {
	refresher, ok := oidcProviderRegistry[claims.Provider]
	if !ok {
		return "", fmt.Errorf("provider '%v' not found in registry", claims.Provider)
	}
	return refresher.RefreshSessionJWT(claims)
}

func (h *authHandler) validateEmailDomain(w http.ResponseWriter, oauthCfg map[string]interface{}, jwtString string) bool {
	if patterns, found := oauthCfg["email_domains"].([]interface{}); found && len(patterns) > 0 {
		tkn, _ := jwt.ParseWithClaims(jwtString, &zrokClaims{}, func(t *jwt.Token) (interface{}, error) {
			return h.signingKey, nil
		})
		claims := tkn.Claims.(*zrokClaims)

		for _, pattern := range patterns {
			if castedPattern, ok := pattern.(string); ok {
				match, err := glob.Compile(castedPattern)
				if err != nil {
					err := fmt.Errorf("invalid email address pattern glob '%v': %v", pattern, err)
					dl.Error(err)
					proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email).WithError(err))
					return false
				}
				if match.Match(claims.Email) {
					return true
				}
			}
		}
		dl.Warnf("unauthorized email '%v'", claims.Email)
		proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedUser(claims.Email))
		return false
	}
	return true
}

func getRefreshInterval(oauthCfg map[string]interface{}) time.Duration {
	if refreshInterval, found := oauthCfg["authorization_check_interval"]; !found {
		dl.Error("missing 'authorization_check_interval', defaulting to 3 hours")
		return 3 * time.Hour
	} else {
		i, err := time.ParseDuration(refreshInterval.(string))
		if err != nil {
			dl.Errorf("invalid refresh interval '%v', defaulting to 3 hours", refreshInterval)
			return 3 * time.Hour
		}
		return i
	}
}
