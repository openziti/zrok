package publicProxy

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/zrok/endpoints"
	"github.com/openziti/zrok/endpoints/publicProxy/healthUi"
	"github.com/openziti/zrok/endpoints/publicProxy/notFoundUi"
	"github.com/openziti/zrok/endpoints/publicProxy/unauthorizedUi"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/sdk"
	"github.com/openziti/zrok/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	zhttp "github.com/zitadel/oidc/v2/pkg/http"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

type httpFrontend struct {
	cfg     *Config
	zCtx    ziti.Context
	handler http.Handler
}

func NewHTTP(cfg *Config) (*httpFrontend, error) {
	env, err := environment.LoadRoot()
	if err != nil {
		return nil, errors.Wrap(err, "error loading environment root")
	}
	zCfgPath, err := env.ZitiIdentityNamed(cfg.Identity)
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
	if err := configureOauthHandlers(context.Background(), cfg, false); err != nil {
		return nil, err
	}
	handler := authHandler(util.NewProxyHandler(proxy), "zrok", cfg, zCtx)
	return &httpFrontend{
		cfg:     cfg,
		zCtx:    zCtx,
		handler: handler,
	}, nil
}

func (self *httpFrontend) Run() error {
	return http.ListenAndServe(self.cfg.Address, self.handler)
}

type zitiDialContext struct {
	ctx ziti.Context
}

func (self *zitiDialContext) Dial(_ context.Context, _ string, addr string) (net.Conn, error) {
	shrToken := strings.Split(addr, ":")[0] // ignore :port (we get passed 'host:port')
	conn, err := self.ctx.Dial(shrToken)
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

func authHandler(handler http.Handler, realm string, pcfg *Config, ctx ziti.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shrToken := resolveService(pcfg.HostMatch, r.Host)
		if shrToken != "" {
			if svc, found := endpoints.GetRefreshedService(shrToken, ctx); found {
				if cfg, found := svc.Config[sdk.ZrokProxyConfig]; found {
					if scheme, found := cfg["auth_scheme"]; found {
						switch scheme {
						case string(sdk.None):
							logrus.Debugf("auth scheme none '%v'", shrToken)
							handler.ServeHTTP(w, r)
							return

						case string(sdk.Basic):
							logrus.Debugf("auth scheme basic '%v", shrToken)
							inUser, inPass, ok := r.BasicAuth()
							if !ok {
								writeUnauthorizedResponse(w, realm)
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
								writeUnauthorizedResponse(w, realm)
								return
							}

							handler.ServeHTTP(w, r)

						case string(sdk.Oauth):
							if oauthCfg, found := cfg["oauth"]; found {
								if provider, found := oauthCfg.(map[string]interface{})["provider"]; found {
									var authCheckInterval time.Duration
									if checkInterval, found := oauthCfg.(map[string]interface{})["authorization_check_interval"]; !found {
										logrus.Errorf("Missing authorization check interval in share config. Defaulting to 3 hours")
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

									cookie, err := r.Cookie("zrok-access")
									if err != nil {
										logrus.Errorf("Unable to get access cookie: %v", err)
										http.Redirect(w, r, fmt.Sprintf("http://%s.%s:28080/%s/login?share=%s&checkInterval=%s", shrToken, pcfg.HostMatch, provider.(string), shrToken, authCheckInterval.String()), http.StatusFound)
										return
									}
									tkn, err := jwt.ParseWithClaims(cookie.Value, &ZrokClaims{}, func(t *jwt.Token) (interface{}, error) {
										if pcfg.Oauth == nil {
											return nil, fmt.Errorf("missing oauth configuration for access point. Unable to parse jwt")
										}
										return []byte(pcfg.Oauth.HashKeyRaw), nil
									})
									if err != nil {
										logrus.Errorf("Unable to parse JWT: %v", err)
										http.Redirect(w, r, fmt.Sprintf("http://%s.%s:28080/%s/login?share=%s&checkInterval=%s", shrToken, pcfg.HostMatch, provider.(string), shrToken, authCheckInterval.String()), http.StatusFound)
										return
									}
									claims := tkn.Claims.(*ZrokClaims)
									if claims.Provider != provider {
										logrus.Error("Provider mismatch. Redoing auth flow")
										http.Redirect(w, r, fmt.Sprintf("http://%s.%s:28080/%s/login?share=%s&checkInterval=%s", shrToken, pcfg.HostMatch, provider.(string), shrToken, authCheckInterval.String()), http.StatusFound)
										return
									}
									if claims.AuthorizationCheckInterval != authCheckInterval {
										logrus.Error("Authorization check interval mismatch. Redoing auth flow")
										http.Redirect(w, r, fmt.Sprintf("http://%s.%s:28080/%s/login?share=%s&checkInterval=%s", shrToken, pcfg.HostMatch, provider.(string), shrToken, authCheckInterval.String()), http.StatusFound)
										return
									}
									if validDomains, found := oauthCfg.(map[string]interface{})["email_domains"]; found {
										if castedDomains, ok := validDomains.([]interface{}); !ok {
											logrus.Error("Invalid format for valid email domains")
											return
										} else {
											if len(castedDomains) > 0 {
												found := false
												for _, domain := range castedDomains {
													if strings.HasSuffix(claims.Email, domain.(string)) {
														found = true
														break
													}
												}
												if !found {
													logrus.Warnf("Email not a valid domain")
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
							writeUnauthorizedResponse(w, realm)
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

func configureOauthHandlers(ctx context.Context, cfg *Config, tls bool) error {
	if cfg.Oauth == nil {
		logrus.Info("No oauth config for access point. Skipping spin up.")
		return nil
	}
	if err := configureGoogleOauth(cfg.Oauth, tls); err != nil {
		return err
	}
	if err := configureGithubOauth(cfg.Oauth, tls); err != nil {
		return err
	}
	zhttp.StartServer(ctx, fmt.Sprintf("%s:%d", strings.Split(cfg.Address, ":")[0], cfg.Oauth.Port))
	return nil
}

type ZrokClaims struct {
	Email                      string        `json:"email"`
	AccessToken                string        `json:"accessToken"`
	Provider                   string        `json:"provider"`
	AuthorizationCheckInterval time.Duration `json:"authorizationCheckInterval"`
	jwt.RegisteredClaims
}

func SetZrokCookie(w http.ResponseWriter, email, accessToken, provider string, checkInterval time.Duration, key []byte) {
	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, ZrokClaims{
		Email:                      email,
		AccessToken:                accessToken,
		Provider:                   provider,
		AuthorizationCheckInterval: checkInterval,
	})
	sTkn, err := tkn.SignedString(key)
	if err != nil {
		http.Error(w, fmt.Sprintf("After signing cookie token: %v", err.Error()), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "zrok-access",
		Value:   sTkn,
		MaxAge:  int(checkInterval.Seconds()),
		Domain:  "localzrok.io",
		Path:    "/",
		Expires: time.Now().Add(checkInterval),
		//Secure:  true, //When tls gets added have this be configured on if tls
	})
}

func writeUnauthorizedResponse(w http.ResponseWriter, realm string) {
	w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
	w.WriteHeader(401)
	w.Write([]byte("No Authorization\n"))
}

func resolveService(hostMatch string, host string) string {
	logrus.Debugf("host = '%v'", host)
	if hostMatch == "" || strings.Contains(host, hostMatch) {
		tokens := strings.Split(host, ".")
		if len(tokens) > 0 {
			return tokens[0]
		}
	}
	return ""
}
