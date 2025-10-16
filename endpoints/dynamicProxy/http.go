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

	"github.com/michaelquigley/df/da"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/zrok/endpoints"
	"github.com/openziti/zrok/endpoints/proxyUi"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/openziti/zrok/util"
	"github.com/pkg/errors"
)

type httpListener struct {
	cfg        *config
	zCtx       ziti.Context
	handler    http.Handler
	mappings   *mappings
	signingKey []byte
	zTransport *http.Transport
	server     *http.Server
}

func buildHttpListener(app *da.Application[*config]) error {
	hl, err := newHttpListener(app.Cfg)
	if err != nil {
		return err
	}
	da.Set(app.C, hl)
	return nil
}

func newHttpListener(cfg *config) (*httpListener, error) {
	var signingKey []byte
	var err error
	if cfg.Oauth != nil {
		signingKey, err = endpoints.DeriveKey(cfg.Oauth.SigningKey, 32)
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

	if err := configureOauth(cfg, cfg.Tls != nil); err != nil {
		return nil, err
	}

	return &httpListener{
		cfg:        cfg,
		zCtx:       zCtx,
		handler:    nil, // set in Link()
		signingKey: signingKey,
		zTransport: zTransport,
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

func newServiceProxy(cfg *config, ctx ziti.Context, mappings *mappings) (*httputil.ReverseProxy, error) {
	proxy := hostTargetReverseProxy(cfg, ctx, mappings)
	director := proxy.Director
	proxy.Director = func(req *http.Request) {
		director(req)
		req.Header.Set("X-Proxy", "zrok")
	}
	proxy.ModifyResponse = func(resp *http.Response) error {
		return nil
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		dl.Errorf("error proxying: %v", err)
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

func hostTargetReverseProxy(cfg *config, ctx ziti.Context, mappings *mappings) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		targetMapping, found := mappings.getMapping(req.Host)
		if found {
			if svc, found := endpoints.GetRefreshedService(targetMapping.ShareToken, ctx); found {
				if cfg, found := svc.Config[sdk.ZrokProxyConfig]; found {
					dl.Debugf("auth model: %v", cfg)
				} else {
					dl.Warn("no config!")
				}
				if target, err := url.Parse(fmt.Sprintf("http://%v", targetMapping.ShareToken)); err == nil {
					dl.Infof("[%v] -> %v", target, req.URL)

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
					dl.Errorf("error proxying: %v", err)
				}
			}
		}
	}
	return &httputil.ReverseProxy{Director: director}
}

func shareHandler(handler http.Handler, cfg *config, signingKey []byte, ctx ziti.Context, mappings *mappings) http.HandlerFunc {
	auth := newAuthHandler(cfg, signingKey)

	return func(w http.ResponseWriter, r *http.Request) {
		mapping, found := mappings.getMapping(r.Host)
		if !found {
			dl.Debugf("mapping not found for '%v'", r.Host)
			proxyUi.WriteNotFound(w, proxyUi.NotFoundData(r.Host))
			return
		}

		svc, found := endpoints.GetRefreshedService(mapping.ShareToken, ctx)
		if !found {
			dl.Warnf("%v -> service '%v' not found", r.RemoteAddr, mapping.ShareToken)
			proxyUi.WriteNotFound(w, proxyUi.NotFoundData(mapping.ShareToken))
			return
		}

		svcCfg, found := svc.Config[sdk.ZrokProxyConfig]
		if !found {
			dl.Warnf("%v -> no proxy config for '%v'", r.RemoteAddr, mapping.ShareToken)
			proxyUi.WriteNotFound(w, proxyUi.NotFoundData(mapping.ShareToken))
			return
		}

		if handleInterstitial(w, r, cfg, svcCfg) {
			return
		}

		authScheme, found := svcCfg["auth_scheme"]
		if !found {
			dl.Warnf("%v -> no auth scheme for '%v'", r.RemoteAddr, mapping.ShareToken)
			proxyUi.WriteNotFound(w, proxyUi.NotFoundData(mapping.ShareToken))
			return
		}

		switch authScheme {
		case string(sdk.None):
			dl.Debugf("auth scheme none '%v'", mapping.ShareToken)
			filterSessionCookies(w, r, cfg)
			handler.ServeHTTP(w, r)

		case string(sdk.Basic):
			dl.Debugf("auth scheme basic '%v'", mapping.ShareToken)
			if auth.handleBasicAuth(w, r, svcCfg, mapping.ShareToken) {
				filterSessionCookies(w, r, cfg)
				handler.ServeHTTP(w, r)
			}

		case string(sdk.Oauth):
			dl.Debugf("auth scheme oauth '%v'", mapping.ShareToken)
			if auth.handleOAuth(w, r, svcCfg, mapping.ShareToken) {
				handler.ServeHTTP(w, r)
			}

		default:
			err := fmt.Errorf("invalid auth scheme '%v'", authScheme)
			dl.Error(err)
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
					dl.Debugf("forcing interstitial for '%v'", r.URL)
					proxyUi.WriteInterstitialAnnounce(w, pcfg.Interstitial.HtmlPath)
					return true
				}
			}
		}
	}
	return false
}

func (l *httpListener) initializeHandler() error {
	proxy, err := newServiceProxy(l.cfg, l.zCtx, l.mappings)
	if err != nil {
		return err
	}
	proxy.Transport = l.zTransport

	l.handler = shareHandler(util.NewRequestsWrapper(proxy), l.cfg, l.signingKey, l.zCtx, l.mappings)
	return nil
}

func (l *httpListener) Link(c *da.Container) error {
	var found bool
	l.mappings, found = da.Get[*mappings](c)
	if !found {
		return errors.New("mapping not found")
	}

	// now that we have mappings, create the proxy and handler
	if err := l.initializeHandler(); err != nil {
		return err
	}

	return nil
}

func (l *httpListener) Start() error {
	l.server = &http.Server{
		Addr:    l.cfg.BindAddress,
		Handler: l.handler,
	}

	if l.cfg.Tls != nil {
		go func() {
			if err := l.server.ListenAndServeTLS(l.cfg.Tls.CertPath, l.cfg.Tls.KeyPath); err != nil && !errors.Is(err, http.ErrServerClosed) {
				dl.Error(err)
			}
		}()
		dl.Infof("started TLS listener")

	} else {
		go func() {
			if err := l.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				dl.Error(err)
			}
		}()
		dl.Infof("started HTTP listener")
	}
	return nil
}

func (l *httpListener) Stop() error {
	if l.server == nil {
		return nil
	}

	// create a timeout context for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dl.Info("shutting down HTTP listener")
	if err := l.server.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "error shutting down HTTP server")
	}

	return nil
}
