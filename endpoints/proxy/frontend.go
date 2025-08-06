package proxy

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/zrok/endpoints"
	"github.com/openziti/zrok/endpoints/proxyUi"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/openziti/zrok/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type FrontendConfig struct {
	IdentityName    string
	ShrToken        string
	Address         string
	ResponseHeaders []string
	TemplatePath    string
	Tls             *endpoints.TlsConfig
	RequestsChan    chan *endpoints.Request
	SuperNetwork    bool
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
	env, err := environment.LoadRoot()
	if err != nil {
		return nil, errors.Wrap(err, "error loading environment root")
	}
	zCfgPath, err := env.ZitiIdentityNamed(cfg.IdentityName)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting ziti identity '%v' from environment", cfg.IdentityName)
	}
	zCfg, err := ziti.NewConfigFromFile(zCfgPath)
	if err != nil {
		return nil, errors.Wrap(err, "error loading config")
	}
	zCfg.ConfigTypes = []string{sdk.ZrokProxyConfig}
	if cfg.SuperNetwork {
		util.EnableSuperNetwork(zCfg)
	}
	zCtx, err := ziti.NewContext(zCfg)
	if err != nil {
		return nil, errors.Wrap(err, "error loading ziti context")
	}
	zDialCtx := zitiDialContext{ctx: zCtx, shrToken: cfg.ShrToken}
	zTransport := http.DefaultTransport.(*http.Transport).Clone()
	zTransport.DialContext = zDialCtx.Dial

	proxy, err := newServiceProxy(cfg, zCtx)
	if err != nil {
		return nil, err
	}
	proxy.Transport = zTransport

	handler := authHandler(cfg.ShrToken, util.NewRequestsWrapper(proxy), "zrok", cfg, zCtx)
	return &Frontend{
		cfg:     cfg,
		zCtx:    zCtx,
		handler: handler,
	}, nil
}

func (h *Frontend) Run() error {
	if h.cfg.Tls != nil {
		return http.ListenAndServeTLS(h.cfg.Address, h.cfg.Tls.CertPath, h.cfg.Tls.KeyPath, h.handler)
	}
	return http.ListenAndServe(h.cfg.Address, h.handler)
}

type zitiDialContext struct {
	ctx      ziti.Context
	shrToken string
}

func (zdc *zitiDialContext) Dial(_ context.Context, _ string, addr string) (net.Conn, error) {
	conn, err := zdc.ctx.DialWithOptions(zdc.shrToken, &ziti.DialOptions{ConnectTimeout: 30 * time.Second})
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
		for _, responseHeader := range cfg.ResponseHeaders {
			tokens := strings.Split(responseHeader, ":")
			if len(tokens) == 2 {
				resp.Header.Set(strings.TrimSpace(tokens[0]), strings.TrimSpace(tokens[1]))
			} else {
				logrus.Errorf("invalid response header '%v' (expecting header:value", responseHeader)
			}
		}
		return nil
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logrus.Errorf("error proxying: %v", err)
		proxyUi.WriteBadGateway(
			w, proxyUi.TemplateData(
				"bad gateway!",
				fmt.Sprintf("bad gateway for share <code>%v</code>!", cfg.ShrToken),
			),
			cfg.TemplatePath,
		)
	}
	return proxy, nil
}

func serviceTargetProxy(cfg *FrontendConfig, ctx ziti.Context) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		targetShrToken := cfg.ShrToken
		if svc, found := endpoints.GetRefreshedService(targetShrToken, ctx); found {
			if cfg, found := svc.Config[sdk.ZrokProxyConfig]; found {
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
			if proxyCfg, found := svc.Config[sdk.ZrokProxyConfig]; found {
				if scheme, found := proxyCfg["auth_scheme"]; found {
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
						if v, found := proxyCfg["basic_auth"]; found {
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
					proxyUi.WriteNotFound(w, proxyUi.NotFoundData(shrToken), cfg.TemplatePath)
				}
			} else {
				logrus.Warnf("%v -> no proxy config for '%v'", r.RemoteAddr, shrToken)
				proxyUi.WriteNotFound(w, proxyUi.NotFoundData(shrToken), cfg.TemplatePath)
			}
		} else {
			logrus.Warnf("%v -> service '%v' not found", r.RemoteAddr, shrToken)
			proxyUi.WriteNotFound(w, proxyUi.NotFoundData(shrToken), cfg.TemplatePath)
		}
	}
}

func writeUnauthorizedResponse(w http.ResponseWriter, realm string) {
	w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
	w.WriteHeader(401)
	w.Write([]byte("No Authorization\n"))
}
