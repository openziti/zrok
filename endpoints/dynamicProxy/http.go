package dynamicProxy

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/michaelquigley/df"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/zrok/endpoints"
	"github.com/openziti/zrok/endpoints/proxyUi"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/openziti/zrok/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type httpListener struct {
	cfg      *config
	zCtx     ziti.Context
	handler  http.Handler
	mappings *mappings
}

func buildHttpListener(app *df.Application[*config]) error {
	hl, err := newHttpListener(app.Cfg)
	if err != nil {
		return err
	}
	df.Set(app.C, hl)
	return nil
}

func newHttpListener(cfg *config) (*httpListener, error) {
	var signingKey []byte
	var err error
	if cfg.Oauth != nil {
		signingKey, err = deriveKey(cfg.Oauth.SigningKey, 32)
		if err != nil {
			return nil, err
		}
	}

	if cfg.TemplatePath != "" {
		if err := proxyUi.ReplaceTemplate(cfg.TemplatePath); err != nil {
			return nil, err
		}
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
	superNetwork, _ := root.SuperNetwork()
	if superNetwork {
		util.EnableSuperNetwork(zCfg)
	}
	zTransport := http.DefaultTransport.(*http.Transport).Clone()
	zTransport.DialContext = zDialCtx.Dial

	proxy, err := newServiceProxy(cfg, zCtx)
	if err != nil {
		return nil, err
	}
	proxy.Transport = zTransport
	if err := configureOauth(context.Background(), cfg, cfg.Tls != nil); err != nil {
		return nil, err
	}
	handler := shareHandler(util.NewRequestsWrapper(proxy), cfg, signingKey, zCtx)
	return &httpListener{
		cfg:     cfg,
		zCtx:    zCtx,
		handler: handler,
	}, nil
}

type zitiDialContext struct {
	ctx ziti.Context
}

func (c *zitiDialContext) Dial(_ context.Context, _ string, addr string) (net.Conn, error) {
	shrToken := strings.Split(addr, ":")[0] // ignore :port (we get passed 'host:port')
	conn, err := c.ctx.DialWithOptions(shrToken, &ziti.DialOptions{ConnectTimeout: 30 * time.Second})
	if err != nil {
		return conn, err
	}
	return conn, nil
}

func newServiceProxy(cfg *config, ctx ziti.Context) (*httputil.ReverseProxy, error) {
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
		proxyUi.WriteBadGateway(
			w,
			proxyUi.RequiredData(
				"bad gateway!",
				"bad gateway!",
			),
		)
	}
	return proxy, nil
}

func hostTargetReverseProxy(cfg *config, ctx ziti.Context) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		targetShrToken := "todo-get-from-mapping"
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

func shareHandler(handler http.Handler, cfg *config, signingKey []byte, ctx ziti.Context) http.HandlerFunc {
	auth := newAuthHandler(cfg, signingKey)

	return func(w http.ResponseWriter, r *http.Request) {
		shrToken := "todo-get-from-mapping"
		if shrToken == "" {
			logrus.Debugf("host '%v' did not match host match, returning health check", r.Host)
			proxyUi.WriteHealthOk(w)
			return
		}

		svc, found := endpoints.GetRefreshedService(shrToken, ctx)
		if !found {
			logrus.Warnf("%v -> service '%v' not found", r.RemoteAddr, shrToken)
			proxyUi.WriteNotFound(w, proxyUi.NotFoundData(shrToken))
			return
		}

		svcCfg, found := svc.Config[sdk.ZrokProxyConfig]
		if !found {
			logrus.Warnf("%v -> no proxy config for '%v'", r.RemoteAddr, shrToken)
			proxyUi.WriteNotFound(w, proxyUi.NotFoundData(shrToken))
			return
		}

		if handleInterstitial(w, r, cfg, svcCfg) {
			return
		}

		authScheme, found := svcCfg["auth_scheme"]
		if !found {
			logrus.Warnf("%v -> no auth scheme for '%v'", r.RemoteAddr, shrToken)
			proxyUi.WriteNotFound(w, proxyUi.NotFoundData(shrToken))
			return
		}

		switch authScheme {
		case string(sdk.None):
			logrus.Debugf("auth scheme none '%v'", shrToken)
			filterSessionCookies(w, r, cfg)
			handler.ServeHTTP(w, r)

		case string(sdk.Basic):
			logrus.Debugf("auth scheme basic '%v'", shrToken)
			if auth.handleBasicAuth(w, r, svcCfg, shrToken) {
				filterSessionCookies(w, r, cfg)
				handler.ServeHTTP(w, r)
			}

		case string(sdk.Oauth):
			logrus.Debugf("auth scheme oauth '%v'", shrToken)
			if auth.handleOAuth(w, r, svcCfg, shrToken) {
				handler.ServeHTTP(w, r)
			}

		default:
			err := fmt.Errorf("invalid auth scheme '%v'", authScheme)
			logrus.Error(err)
			proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData().WithError(err))
		}
	}
}

func handleInterstitial(w http.ResponseWriter, r *http.Request, pcfg *config, cfg map[string]interface{}) bool {
	if r.Method == http.MethodOptions || pcfg.Interstitial == nil || !pcfg.Interstitial.Enabled {
		return false
	}

	sendInterstitial := true
	if len(pcfg.Interstitial.UserAgentPrefixes) > 0 {
		ua := r.Header.Get("User-Agent")
		for _, prefix := range pcfg.Interstitial.UserAgentPrefixes {
			if strings.HasPrefix(ua, prefix) {
				sendInterstitial = true
				break
			}
		}
		sendInterstitial = false
	}

	if sendInterstitial {
		if v, istlFound := cfg["interstitial"]; istlFound {
			if istlEnabled, ok := v.(bool); ok && istlEnabled {
				skip := r.Header.Get("skip_zrok_interstitial")
				_, zrokOkErr := r.Cookie("zrok_interstitial")
				if skip == "" && zrokOkErr != nil {
					logrus.Debugf("forcing interstitial for '%v'", r.URL)
					proxyUi.WriteInterstitialAnnounce(w, pcfg.Interstitial.HtmlPath)
					return true
				}
			}
		}
	}
	return false
}

func (l *httpListener) Link(c *df.Container) error {
	var found bool
	l.mappings, found = df.Get[*mappings](c)
	if !found {
		return errors.New("mapping not found")
	}
	return nil
}
