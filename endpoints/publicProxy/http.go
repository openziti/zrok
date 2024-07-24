package publicProxy

import (
	"context"
	"crypto/md5"
	"fmt"
	"github.com/gobwas/glob"
	"github.com/golang-jwt/jwt/v5"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/zrok/endpoints"
	"github.com/openziti/zrok/endpoints/publicProxy/healthUi"
	"github.com/openziti/zrok/endpoints/publicProxy/interstitialUi"
	"github.com/openziti/zrok/endpoints/publicProxy/notFoundUi"
	"github.com/openziti/zrok/endpoints/publicProxy/unauthorizedUi"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/openziti/zrok/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

type HttpFrontend struct {
	cfg     *Config
	zCtx    ziti.Context
	handler http.Handler
}

func NewHTTP(cfg *Config) (*HttpFrontend, error) {
	var key []byte
	if cfg.Oauth != nil {
		hash := md5.New()
		n, err := hash.Write([]byte(cfg.Oauth.HashKey))
		if err != nil {
			return nil, err
		}
		if n != len(cfg.Oauth.HashKey) {
			return nil, errors.New("short hash")
		}
		key = hash.Sum(nil)
	}

	root, err := environment.LoadRoot()
	if err != nil {
		return nil, errors.Wrap(err, "error loading environment root")
	}
	zCfgPath, err := root.ZitiIdentityNamed(cfg.Identity)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting ziti identity '%v' from environment", cfg.Identity)
	}
	zCfg, err := ziti.NewConfigFromFile(zCfgPath)
	if err != nil {
		return nil, errors.Wrap(err, "error loading config")
	}
	zCfg.ConfigTypes = []string{sdk.ZrokProxyConfig}
	zCtx, err := ziti.NewContext(zCfg)
	if err != nil {
		return nil, errors.Wrap(err, "error loading ziti context")
	}
	zDialCtx := zitiDialContext{ctx: zCtx}
	zTransport := http.DefaultTransport.(*http.Transport).Clone()
	zTransport.DialContext = zDialCtx.Dial

	proxy, err := newServiceProxy(cfg, zCtx)
	if err != nil {
		return nil, err
	}
	proxy.Transport = zTransport
	if err := configureOauthHandlers(context.Background(), cfg, cfg.Tls != nil); err != nil {
		return nil, err
	}
	handler := shareHandler(util.NewRequestsWrapper(proxy), cfg, key, zCtx)
	return &HttpFrontend{
		cfg:     cfg,
		zCtx:    zCtx,
		handler: handler,
	}, nil
}

func (f *HttpFrontend) Run() error {
	if f.cfg.Tls != nil {
		return http.ListenAndServeTLS(f.cfg.Address, f.cfg.Tls.CertPath, f.cfg.Tls.KeyPath, f.handler)
	}
	return http.ListenAndServe(f.cfg.Address, f.handler)
}

type zitiDialContext struct {
	ctx ziti.Context
}

func (c *zitiDialContext) Dial(_ context.Context, _ string, addr string) (net.Conn, error) {
	shrToken := strings.Split(addr, ":")[0] // ignore :port (we get passed 'host:port')
	conn, err := c.ctx.Dial(shrToken)
	if err != nil {
		return conn, err
	}
	return conn, nil
}

func newServiceProxy(cfg *Config, ctx ziti.Context) (*httputil.ReverseProxy, error) {
	proxy := hostTargetReverseProxy(cfg, ctx)
	director := proxy.Director
	proxy.Director = func(req *http.Request) {
		director(req)
		req.Header.Set("X-Proxy", "zrok")
	}
	proxy.ModifyResponse = func(resp *http.Response) error {
		return nil
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logrus.Errorf("error proxying: %v", err)
		notFoundUi.WriteNotFound(w)
	}
	return proxy, nil
}

func hostTargetReverseProxy(cfg *Config, ctx ziti.Context) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		targetShrToken := resolveService(cfg.HostMatch, req.Host)
		if svc, found := endpoints.GetRefreshedService(targetShrToken, ctx); found {
			if cfg, found := svc.Config[sdk.ZrokProxyConfig]; found {
				logrus.Debugf("auth model: %v", cfg)
			} else {
				logrus.Warn("no config!")
			}
			if target, err := url.Parse(fmt.Sprintf("http://%v", targetShrToken)); err == nil {
				logrus.Infof("[%v] -> %v", targetShrToken, req.URL)

				targetQuery := target.RawQuery
				req.URL.Scheme = target.Scheme
				req.URL.Host = target.Host
				req.URL.Path, req.URL.RawPath = endpoints.JoinURLPath(target, req.URL)
				if targetQuery == "" || req.URL.RawQuery == "" {
					req.URL.RawQuery = targetQuery + req.URL.RawQuery
				} else {
					req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
				}
				if _, ok := req.Header["User-Agent"]; !ok {
					// explicitly disable User-Agent so it's not set to default value
					req.Header.Set("User-Agent", "")
				}
			} else {
				logrus.Errorf("error proxying: %v", err)
			}
		}
	}
	return &httputil.ReverseProxy{Director: director}
}

func shareHandler(handler http.Handler, pcfg *Config, key []byte, ctx ziti.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shrToken := resolveService(pcfg.HostMatch, r.Host)
		if shrToken != "" {
			if svc, found := endpoints.GetRefreshedService(shrToken, ctx); found {
				if cfg, found := svc.Config[sdk.ZrokProxyConfig]; found {
					if pcfg.Interstitial {
						if v, istlFound := cfg["interstitial"]; istlFound {
							if istlEnabled, ok := v.(bool); ok && istlEnabled {
								skip := r.Header.Get("skip_zrok_interstitial")
								_, zrokOkErr := r.Cookie("zrok_interstitial")
								if skip == "" && zrokOkErr != nil {
									logrus.Debugf("forcing interstitial for '%v'", r.URL)
									interstitialUi.WriteInterstitialAnnounce(w)
									return
								}
							}
						}
					}

					if scheme, found := cfg["auth_scheme"]; found {
						switch scheme {
						case string(sdk.None):
							logrus.Debugf("auth scheme none '%v'", shrToken)
							// ensure cookies from other shares are not sent to this share, in case it's malicious
							deleteZrokCookies(w, r)
							handler.ServeHTTP(w, r)
							return

						case string(sdk.Basic):
							logrus.Debugf("auth scheme basic '%v", shrToken)
							inUser, inPass, ok := r.BasicAuth()
							if !ok {
								basicAuthRequired(w, shrToken)
								return
							}
							authed := false
							if v, found := cfg["basic_auth"]; found {
								if basicAuth, ok := v.(map[string]interface{}); ok {
									if v, found := basicAuth["users"]; found {
										if arr, ok := v.([]interface{}); ok {
											for _, v := range arr {
												if um, ok := v.(map[string]interface{}); ok {
													username := ""
													if v, found := um["username"]; found {
														if un, ok := v.(string); ok {
															username = un
														}
													}
													password := ""
													if v, found := um["password"]; found {
														if pw, ok := v.(string); ok {
															password = pw
														}
													}
													if username == inUser && password == inPass {
														authed = true
														break
													}
												}
											}
										}
									}
								}
							}

							if !authed {
								basicAuthRequired(w, shrToken)
								return
							}

							// ensure cookies from other shares are not sent to this share, in case it's malicious
							deleteZrokCookies(w, r)
							handler.ServeHTTP(w, r)

						case string(sdk.Oauth):
							logrus.Debugf("auth scheme oauth '%v'", shrToken)

							if oauthCfg, found := cfg["oauth"]; found {
								if provider, found := oauthCfg.(map[string]interface{})["provider"]; found {
									var authCheckInterval time.Duration
									if checkInterval, found := oauthCfg.(map[string]interface{})["authorization_check_interval"]; !found {
										logrus.Errorf("missing authorization check interval in share config. Defaulting to 3 hours")
										authCheckInterval = 3 * time.Hour
									} else {
										i, err := time.ParseDuration(checkInterval.(string))
										if err != nil {
											logrus.Errorf("unable to parse authorization check interval in share config (%v). Defaulting to 3 hours", checkInterval)
											authCheckInterval = 3 * time.Hour
										} else {
											authCheckInterval = i
										}
									}

									target := fmt.Sprintf("%s%s", r.Host, r.URL.Path)

									cookie, err := r.Cookie("zrok-access")
									if err != nil {
										logrus.Errorf("unable to get 'zrok-access' cookie: %v", err)
										oauthLoginRequired(w, r, pcfg.Oauth, provider.(string), target, authCheckInterval)
										return
									}
									tkn, err := jwt.ParseWithClaims(cookie.Value, &ZrokClaims{}, func(t *jwt.Token) (interface{}, error) {
										if pcfg.Oauth == nil {
											return nil, fmt.Errorf("missing oauth configuration for access point; unable to parse jwt")
										}
										return key, nil
									})
									if err != nil {
										logrus.Errorf("unable to parse jwt: %v", err)
										oauthLoginRequired(w, r, pcfg.Oauth, provider.(string), target, authCheckInterval)
										return
									}
									claims := tkn.Claims.(*ZrokClaims)
									if claims.Provider != provider {
										logrus.Error("provider mismatch; restarting auth flow")
										oauthLoginRequired(w, r, pcfg.Oauth, provider.(string), target, authCheckInterval)
										return
									}
									if claims.AuthorizationCheckInterval != authCheckInterval {
										logrus.Error("authorization check interval mismatch; restarting auth flow")
										oauthLoginRequired(w, r, pcfg.Oauth, provider.(string), target, authCheckInterval)
										return
									}
									if claims.Audience != r.Host {
										logrus.Errorf("audience claim '%s' does not match requested host '%s'; restarting auth flow", claims.Audience, r.Host)
										oauthLoginRequired(w, r, pcfg.Oauth, provider.(string), target, authCheckInterval)
										return
									}

									if validEmailAddressPatterns, found := oauthCfg.(map[string]interface{})["email_domains"]; found {
										if castedPatterns, ok := validEmailAddressPatterns.([]interface{}); !ok {
											logrus.Error("invalid email pattern array format")
											return
										} else {
											if len(castedPatterns) > 0 {
												found := false
												for _, pattern := range castedPatterns {
													if castedPattern, ok := pattern.(string); ok {
														match, err := glob.Compile(castedPattern)
														if err != nil {
															logrus.Errorf("invalid email address pattern glob '%v': %v", pattern.(string), err)
															unauthorizedUi.WriteUnauthorized(w)
															return
														}
														if match.Match(claims.Email) {
															found = true
															break
														}
													} else {
														logrus.Errorf("invalid email address pattern '%v'", pattern)
														unauthorizedUi.WriteUnauthorized(w)
														return
													}
												}
												if !found {
													logrus.Warnf("unauthorized email '%v' for '%v'", claims.Email, shrToken)
													unauthorizedUi.WriteUnauthorized(w)
													return
												}
											}
										}
									}
									handler.ServeHTTP(w, r)
									return

								} else {
									logrus.Warnf("%v -> no provider for '%v'", r.RemoteAddr, provider)
									notFoundUi.WriteNotFound(w)
								}
							} else {
								logrus.Warnf("%v -> no oauth cfg for '%v'", r.RemoteAddr, shrToken)
								notFoundUi.WriteNotFound(w)
							}
						default:
							logrus.Infof("invalid auth scheme '%v'", scheme)
							basicAuthRequired(w, shrToken)
							return
						}
					} else {
						logrus.Warnf("%v -> no auth scheme for '%v'", r.RemoteAddr, shrToken)
						notFoundUi.WriteNotFound(w)
					}
				} else {
					logrus.Warnf("%v -> no proxy config for '%v'", r.RemoteAddr, shrToken)
					notFoundUi.WriteNotFound(w)
				}
			} else {
				logrus.Warnf("%v -> service '%v' not found", r.RemoteAddr, shrToken)
				notFoundUi.WriteNotFound(w)
			}
		} else {
			logrus.Debugf("host '%v' did not match host match, returning health check", r.Host)
			healthUi.WriteHealthOk(w)
		}
	}
}

type ZrokClaims struct {
	Email                      string        `json:"email"`
	AccessToken                string        `json:"accessToken"`
	Provider                   string        `json:"provider"`
	Audience                   string        `json:"aud"`
	AuthorizationCheckInterval time.Duration `json:"authorizationCheckInterval"`
	jwt.RegisteredClaims
}

func SetZrokCookie(w http.ResponseWriter, cookieDomain, email, accessToken, provider string, checkInterval time.Duration, key []byte, targetHost string) {
	targetHost = strings.TrimSpace(targetHost)
	if targetHost == "" {
		logrus.Error("host claim must not be empty")
		http.Error(w, "host claim must not be empty", http.StatusBadRequest)
		return
	}
	targetHost = strings.Split(targetHost, "/")[0]
	logrus.Debugf("setting zrok-access cookie JWT audience '%s'", targetHost)

	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, ZrokClaims{
		Email:                      email,
		AccessToken:                accessToken,
		Provider:                   provider,
		Audience:                   targetHost,
		AuthorizationCheckInterval: checkInterval,
	})
	sTkn, err := tkn.SignedString(key)
	if err != nil {
		http.Error(w, fmt.Sprintf("after signing cookie token: %v", err.Error()), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "zrok-access",
		Value:   sTkn,
		MaxAge:  int(checkInterval.Seconds()),
		Domain:  cookieDomain,
		Path:    "/",
		Expires: time.Now().Add(checkInterval),
		// Secure:  true, // pending server tls feature https://github.com/openziti/zrok/issues/24
		HttpOnly: true,                 // enabled because zrok frontend is the only intended consumer of this cookie, not client-side scripts
		SameSite: http.SameSiteLaxMode, // explicitly set to the default Lax mode which allows the zrok share to be navigated to from another site and receive the cookie
	})
}

func deleteZrokCookies(w http.ResponseWriter, r *http.Request) {
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

func basicAuthRequired(w http.ResponseWriter, realm string) {
	w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
	w.WriteHeader(401)
	_, _ = w.Write([]byte("No Authorization\n"))
}

func oauthLoginRequired(w http.ResponseWriter, r *http.Request, cfg *OauthConfig, provider, target string, authCheckInterval time.Duration) {
	http.Redirect(w, r, fmt.Sprintf("%s/%s/login?targethost=%s&checkInterval=%s", cfg.RedirectUrl, provider, url.QueryEscape(target), authCheckInterval.String()), http.StatusFound)
}

func resolveService(hostMatch string, host string) string {
	if hostMatch == "" || strings.Contains(host, hostMatch) {
		tokens := strings.Split(host, ".")
		if len(tokens) > 0 {
			return tokens[0]
		}
	}
	return ""
}
