package dynamicProxy

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gobwas/glob"
	"github.com/golang-jwt/jwt/v5"
	"github.com/michaelquigley/df/dd"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/endpoints/proxyUi"
	"github.com/pkg/errors"
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

func oauthLoginRequired(w http.ResponseWriter, r *http.Request, cfg *oauthConfig, provider, target string, refreshInterval time.Duration) {
	http.Redirect(w, r, fmt.Sprintf("%s/%s/login?targetHost=%s&refreshInterval=%s", cfg.EndpointUrl, provider, url.QueryEscape(target), refreshInterval.String()), http.StatusFound)
}

func oauthRefreshRequired(w http.ResponseWriter, r *http.Request, cfg *oauthConfig, provider, target string) {
	http.Redirect(w, r, fmt.Sprintf("%s/%s/refresh?targetHost=%s", cfg.EndpointUrl, provider, url.QueryEscape(target)), http.StatusFound)
}

func (h *authHandler) handleOAuth(w http.ResponseWriter, r *http.Request, cfg map[string]interface{}, shrToken string) bool {
	oauthCfg, found := cfg["oauth"]
	if !found {
		dl.Warnf("%v -> no oauth cfg for '%v'", r.RemoteAddr, shrToken)
		return false
	}

	oauthMap := oauthCfg.(map[string]interface{})
	provider := oauthMap["provider"].(string)
	refreshInterval := getRefreshInterval(oauthMap)
	target := fmt.Sprintf("%s%s", r.Host, r.URL.Path)

	cookie, err := getSessionCookie(r, h.cfg.Oauth.CookieName)
	if err != nil {
		dl.Errorf("unable to get '%v' cookie: %v", h.cfg.Oauth.CookieName, err)
		oauthLoginRequired(w, r, h.cfg.Oauth, provider, target, refreshInterval)
		return false
	}

	if !h.validateOAuthToken(w, r, cookie, provider, refreshInterval, target) {
		return false
	}

	if !h.validateEmailDomain(w, oauthMap, cookie) {
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
		dl.Errorf("unable to parse jwt: %v", err)
		oauthLoginRequired(w, r, h.cfg.Oauth, provider, target, refreshInterval)
		return false
	}

	claims := tkn.Claims.(*zrokClaims)
	if claims.Provider != provider || claims.RefreshInterval != refreshInterval || claims.TargetHost != r.Host {
		dl.Errorf("token validation failed; restarting auth flow (email: '%v', target: '%v')", claims.Email, target)
		oauthLoginRequired(w, r, h.cfg.Oauth, provider, target, refreshInterval)
		return false
	}

	if time.Now().After(claims.NextRefresh) {
		if claims.SupportsRefresh {
			dl.Infof("oauth session expired; refreshing tokens (email: '%v', target: '%v')", claims.Email, target)
			oauthRefreshRequired(w, r, h.cfg.Oauth, provider, target)
		} else {
			dl.Warnf("oauth session expired; re-login (email: '%v', target: '%v')", claims.Email, target)
			oauthLoginRequired(w, r, h.cfg.Oauth, provider, target, refreshInterval)
		}
		return false
	} else {
		dl.Debugf("%v until next refresh", time.Until(claims.NextRefresh))
	}

	r.Header.Set("zrok-auth-provider", provider)
	r.Header.Set("zrok-auth-email", claims.Email)
	r.Header.Set("zrok-auth-expires", claims.NextRefresh.Format(time.RFC3339))

	return true
}

func (h *authHandler) validateEmailDomain(w http.ResponseWriter, oauthCfg map[string]interface{}, cookie *http.Cookie) bool {
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

func configureOauth(cfg *config, tls bool) error {
	if cfg.Oauth == nil {
		dl.Info("no oauth configuration; skipping oauth handler startup")
		return nil
	}

	if globalOAuthRouter == nil {
		return errors.New("oauth router not initialized")
	}

	// configure providers (they will register themselves with the router)
	for _, v := range cfg.Oauth.Providers {
		if prvCfg, ok := v.(dd.Dynamic); ok {
			switch prvCfg.Type() {
			case "github":
				githubCfg, ok := prvCfg.(*githubConfig)
				if !ok {
					return errors.New("invalid github provider configuration")
				}
				if err := githubCfg.configure(cfg.Oauth, tls); err != nil {
					return err
				}

			case "google":
				googleCfg, ok := prvCfg.(*googleConfig)
				if !ok {
					return errors.New("invalid google provider configuration")
				}
				if err := googleCfg.configure(cfg.Oauth, tls); err != nil {
					return err
				}

			case "oidc":
				oidcCfg, ok := prvCfg.(*oidcConfig)
				if !ok {
					return errors.New("invalid oidc provider configuration")
				}
				if err := oidcCfg.configure(cfg.Oauth, tls); err != nil {
					return err
				}

			default:
				return errors.Errorf("invalid oauth provider type '%v'", prvCfg.Type())
			}
		} else {
			return errors.Errorf("invalid oauth provider configuration; missing 'type'")
		}
	}
	return nil
}
