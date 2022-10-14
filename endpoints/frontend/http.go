package frontend

import (
	"context"
	"fmt"
	"github.com/openziti-test-kitchen/zrok/endpoints/frontend/health_ui"
	"github.com/openziti-test-kitchen/zrok/endpoints/frontend/notfound_ui"
	"github.com/openziti-test-kitchen/zrok/model"
	"github.com/openziti-test-kitchen/zrok/util"
	"github.com/openziti-test-kitchen/zrok/zrokdir"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/config"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type httpListen struct {
	cfg     *Config
	zCtx    ziti.Context
	handler http.Handler
	metrics *metricsAgent
}

func NewHTTP(cfg *Config) (*httpListen, error) {
	ma, err := newMetricsAgent(cfg.Identity)
	if err != nil {
		return nil, err
	}
	go ma.run()

	zCfgPath, err := zrokdir.ZitiIdentityFile(cfg.Identity)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting ziti identity '%v' from zrokdir", cfg.Identity)
	}
	zCfg, err := config.NewFromFile(zCfgPath)
	if err != nil {
		return nil, errors.Wrap(err, "error loading config")
	}
	zCfg.ConfigTypes = []string{model.ZrokProxyConfig}
	zCtx := ziti.NewContextWithConfig(zCfg)
	zDialCtx := zitiDialContext{ctx: zCtx, updates: ma.updates}
	zTransport := http.DefaultTransport.(*http.Transport).Clone()
	zTransport.DialContext = zDialCtx.Dial

	proxy, err := newServiceProxy(cfg, zCtx)
	if err != nil {
		return nil, err
	}
	proxy.Transport = zTransport

	handler := authHandler(util.NewProxyHandler(proxy), "zrok", cfg, zCtx)
	return &httpListen{
		cfg:     cfg,
		zCtx:    zCtx,
		handler: handler,
		metrics: ma,
	}, nil
}

func (self *httpListen) Run() error {
	return http.ListenAndServe(self.cfg.Address, self.handler)
}

type zitiDialContext struct {
	ctx     ziti.Context
	updates chan metricsUpdate
}

func (self *zitiDialContext) Dial(_ context.Context, _ string, addr string) (net.Conn, error) {
	svcName := strings.Split(addr, ":")[0] // ignore :port (we get passed 'host:port')
	conn, err := self.ctx.Dial(svcName)
	if err != nil {
		return conn, err
	}
	return newMetricsConn(svcName, conn, self.updates), nil
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
		notfound_ui.WriteNotFound(w)
	}

	return proxy, nil
}

func hostTargetReverseProxy(cfg *Config, ctx ziti.Context) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		targetSvc := resolveService(cfg.HostMatch, req.Host)
		if svc, found := getRefreshedService(targetSvc, ctx); found {
			if cfg, found := svc.Configs[model.ZrokProxyConfig]; found {
				logrus.Debugf("auth model: %v", cfg)
			} else {
				logrus.Warn("no config!")
			}
			if target, err := url.Parse(fmt.Sprintf("http://%v", targetSvc)); err == nil {
				logrus.Infof("[%v] -> %v", targetSvc, req.URL)

				targetQuery := target.RawQuery
				req.URL.Scheme = target.Scheme
				req.URL.Host = target.Host
				req.URL.Path, req.URL.RawPath = joinURLPath(target, req.URL)
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

func joinURLPath(a, b *url.URL) (path, rawpath string) {
	if a.RawPath == "" && b.RawPath == "" {
		return singleJoiningSlash(a.Path, b.Path), ""
	}
	// Same as singleJoiningSlash, but uses EscapedPath to determine
	// whether a slash should be added
	apath := a.EscapedPath()
	bpath := b.EscapedPath()

	aslash := strings.HasSuffix(apath, "/")
	bslash := strings.HasPrefix(bpath, "/")

	switch {
	case aslash && bslash:
		return a.Path + b.Path[1:], apath + bpath[1:]
	case !aslash && !bslash:
		return a.Path + "/" + b.Path, apath + "/" + bpath
	}
	return a.Path + b.Path, apath + bpath
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func authHandler(handler http.Handler, realm string, cfg *Config, ctx ziti.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		svcName := resolveService(cfg.HostMatch, r.Host)
		if svcName != "" {
			if svc, found := getRefreshedService(svcName, ctx); found {
				if cfg, found := svc.Configs[model.ZrokProxyConfig]; found {
					if scheme, found := cfg["auth_scheme"]; found {
						switch scheme {
						case string(model.None):
							logrus.Debugf("auth scheme none '%v'", svcName)
							handler.ServeHTTP(w, r)
							return

						case string(model.Basic):
							logrus.Debugf("auth scheme basic '%v", svcName)
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

						default:
							logrus.Infof("invalid auth scheme '%v'", scheme)
							writeUnauthorizedResponse(w, realm)
							return
						}
					} else {
						logrus.Warnf("%v -> no auth scheme for '%v'", r.RemoteAddr, svcName)
						notfound_ui.WriteNotFound(w)
					}
				} else {
					logrus.Warnf("%v -> no proxy config for '%v'", r.RemoteAddr, svcName)
					notfound_ui.WriteNotFound(w)
				}
			} else {
				logrus.Warnf("%v -> service '%v' not found", r.RemoteAddr, svcName)
				notfound_ui.WriteNotFound(w)
			}
		} else {
			logrus.Debugf("host '%v' did not match host match, returning health check", r.Host)
			health_ui.WriteHealthOk(w)
		}
	}
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

func getRefreshedService(name string, ctx ziti.Context) (*edge.Service, bool) {
	svc, found := ctx.GetService(name)
	if !found {
		if err := ctx.RefreshServices(); err != nil {
			logrus.Errorf("error refreshing services: %v", err)
			return nil, false
		}
		return ctx.GetService(name)
	}
	return svc, found
}
