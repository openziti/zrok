package proxy

import (
	"context"
	"fmt"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/config"
	"github.com/openziti/zrok/endpoints"
	"github.com/openziti/zrok/endpoints/publicProxyFrontend/notFoundUi"
	"github.com/openziti/zrok/model"
	"github.com/openziti/zrok/util"
	"github.com/openziti/zrok/zrokdir"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type FrontendConfig struct {
	IdentityName string
	ShrToken     string
	Address      string
	RequestsChan chan *endpoints.Request
}

func DefaultFrontendConfig(identityName string) *FrontendConfig {
	return &FrontendConfig{
		IdentityName: identityName,
		Address:      "0.0.0.0:8080",
	}
}

type Frontend struct {
	cfg      *FrontendConfig
	zCtx     ziti.Context
	shrToken string
	handler  http.Handler
}

func NewFrontend(cfg *FrontendConfig) (*Frontend, error) {
	zCfgPath, err := zrokdir.ZitiIdentityFile(cfg.IdentityName)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting ziti identity '%v' from zrokdir", cfg.IdentityName)
	}
	zCfg, err := config.NewFromFile(zCfgPath)
	if err != nil {
		return nil, errors.Wrap(err, "error loading config")
	}
	zCfg.ConfigTypes = []string{model.ZrokProxyConfig}
	zCtx := ziti.NewContextWithConfig(zCfg)
	zDialCtx := zitiDialContext{ctx: zCtx, shrToken: cfg.ShrToken}
	zTransport := http.DefaultTransport.(*http.Transport).Clone()
	zTransport.DialContext = zDialCtx.Dial

	proxy, err := newServiceProxy(cfg, zCtx)
	if err != nil {
		return nil, err
	}
	proxy.Transport = zTransport

	handler := authHandler(cfg.ShrToken, util.NewProxyHandler(proxy), "zrok", cfg, zCtx)
	return &Frontend{
		cfg:     cfg,
		zCtx:    zCtx,
		handler: handler,
	}, nil
}

func (h *Frontend) Run() error {
	return http.ListenAndServe(h.cfg.Address, h.handler)
}

type zitiDialContext struct {
	ctx      ziti.Context
	shrToken string
}

func (zdc *zitiDialContext) Dial(_ context.Context, _ string, addr string) (net.Conn, error) {
	conn, err := zdc.ctx.Dial(zdc.shrToken)
	if err != nil {
		return conn, err
	}
	return conn, nil
}

func newServiceProxy(cfg *FrontendConfig, ctx ziti.Context) (*httputil.ReverseProxy, error) {
	proxy := serviceTargetProxy(cfg, ctx)
	director := proxy.Director
	proxy.Director = func(req *http.Request) {
		if cfg.RequestsChan != nil {
			cfg.RequestsChan <- &endpoints.Request{
				Stamp:      time.Now(),
				RemoteAddr: fmt.Sprintf("%v", req.Header["X-Real-Ip"]),
				Method:     req.Method,
				Path:       req.URL.String(),
			}
		}
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

func serviceTargetProxy(cfg *FrontendConfig, ctx ziti.Context) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		targetShrToken := cfg.ShrToken
		if svc, found := endpoints.GetRefreshedService(targetShrToken, ctx); found {
			if cfg, found := svc.Configs[model.ZrokProxyConfig]; found {
				logrus.Debugf("auth model: %v", cfg)
			} else {
				logrus.Warn("no config!")
			}
			if target, err := url.Parse(fmt.Sprintf("http://%v", targetShrToken)); err == nil {
				logrus.Debugf("[%v] -> %v", targetShrToken, req.URL)

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

func authHandler(shrToken string, handler http.Handler, realm string, cfg *FrontendConfig, ctx ziti.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if svc, found := endpoints.GetRefreshedService(shrToken, ctx); found {
			if cfg, found := svc.Configs[model.ZrokProxyConfig]; found {
				if scheme, found := cfg["auth_scheme"]; found {
					switch scheme {
					case string(model.None):
						logrus.Debugf("auth scheme none '%v'", shrToken)
						handler.ServeHTTP(w, r)
						return

					case string(model.Basic):
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
	}
}

func writeUnauthorizedResponse(w http.ResponseWriter, realm string) {
	w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
	w.WriteHeader(401)
	w.Write([]byte("No Authorization\n"))
}
