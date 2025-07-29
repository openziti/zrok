package publicProxy

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gobwas/glob"
	"github.com/golang-jwt/jwt/v5"
	"github.com/openziti/zrok/endpoints/publicProxy/unauthorizedUi"
	"github.com/sirupsen/logrus"
)

type ZrokClaims struct {
	Email                      string        `json:"email"`
	AccessToken                string        `json:"accessToken"`
	Provider                   string        `json:"provider"`
	Audience                   string        `json:"aud"`
	AuthorizationCheckInterval time.Duration `json:"authorizationCheckInterval"`
	jwt.RegisteredClaims
}

func (c *ZrokClaims) GetAudience() (jwt.ClaimStrings, error) {
	return jwt.ClaimStrings{c.Audience}, nil
}

type authHandler struct {
	cfg     *Config
	key     []byte
	handler http.Handler
}

func newAuthHandler(cfg *Config, key []byte, handler http.Handler) *authHandler {
	return &authHandler{
		cfg:     cfg,
		key:     key,
		handler: handler,
	}
}

func (h *authHandler) handleBasicAuth(w http.ResponseWriter, r *http.Request, cfg map[string]interface{}, shrToken string) bool {
	inUser, inPass, ok := r.BasicAuth()
	if !ok {
		basicAuthRequired(w, shrToken)
		return false
	}

	if v, found := cfg["basic_auth"]; found {
		if basicAuth, ok := v.(map[string]interface{}); ok {
			if users, found := basicAuth["users"].([]interface{}); found {
				for _, v := range users {
					if um, ok := v.(map[string]interface{}); ok {
						if um["username"] == inUser && um["password"] == inPass {
							return true
						}
					}
				}
			}
		}
	}

	basicAuthRequired(w, shrToken)
	return false
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
	tkn, err := jwt.ParseWithClaims(cookie.Value, &ZrokClaims{}, func(t *jwt.Token) (interface{}, error) {
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

	claims := tkn.Claims.(*ZrokClaims)
	if claims.Provider != provider ||
		claims.AuthorizationCheckInterval != authCheckInterval ||
		claims.Audience != r.Host {
		logrus.Error("token validation failed; restarting auth flow")
		oauthLoginRequired(w, r, h.cfg.Oauth, provider, target, authCheckInterval)
		return false
	}

	return true
}

func (h *authHandler) validateEmailDomain(w http.ResponseWriter, oauthCfg map[string]interface{}, cookie *http.Cookie) bool {
	if patterns, found := oauthCfg["email_domains"].([]interface{}); found && len(patterns) > 0 {
		tkn, _ := jwt.ParseWithClaims(cookie.Value, &ZrokClaims{}, func(t *jwt.Token) (interface{}, error) {
			return h.key, nil
		})
		claims := tkn.Claims.(*ZrokClaims)

		for _, pattern := range patterns {
			if castedPattern, ok := pattern.(string); ok {
				match, err := glob.Compile(castedPattern)
				if err != nil {
					logrus.Errorf("invalid email address pattern glob '%v': %v", pattern, err)
					unauthorizedUi.WriteUnauthorized(w)
					return false
				}
				if match.Match(claims.Email) {
					return true
				}
			}
		}
		logrus.Warnf("unauthorized email '%v'", claims.Email)
		unauthorizedUi.WriteUnauthorized(w)
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

func basicAuthRequired(w http.ResponseWriter, realm string) {
	w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
	w.WriteHeader(401)
	_, _ = w.Write([]byte("No Authorization\n"))
}

func oauthLoginRequired(w http.ResponseWriter, r *http.Request, cfg *OauthConfig, provider, target string, authCheckInterval time.Duration) {
	http.Redirect(w, r, fmt.Sprintf("%s/%s/login?targethost=%s&checkInterval=%s", cfg.RedirectUrl, provider, url.QueryEscape(target), authCheckInterval.String()), http.StatusFound)
}
