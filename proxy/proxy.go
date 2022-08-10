package proxy

import (
	"context"
	"fmt"
	"github.com/openziti-test-kitchen/zrok/util"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type Config struct {
	IdentityPath string
	Address      string
}

func Run(cfg *Config) error {
	zCfg, err := config.NewFromFile(cfg.IdentityPath)
	if err != nil {
		return errors.Wrap(err, "error loading config")
	}
	zCtx := ziti.NewContextWithConfig(zCfg)
	zDialCtx := ZitiDialContext{Context: zCtx}
	zTransport := http.DefaultTransport.(*http.Transport).Clone()
	zTransport.DialContext = zDialCtx.Dial

	proxy, err := NewServiceProxy(&resolver{})
	if err != nil {
		return err
	}
	proxy.Transport = zTransport
	return http.ListenAndServe(cfg.Address, util.NewProxyHandler(proxy))
}

type resolver struct{}

func (r *resolver) Service(host string) string {
	logrus.Debugf("host = '%v'", host)
	tokens := strings.Split(host, ".")
	if len(tokens) > 0 {
		return tokens[0]
	}
	return "zrok"
}

type ZitiDialContext struct {
	Context ziti.Context
}

func (self *ZitiDialContext) Dial(_ context.Context, _ string, addr string) (net.Conn, error) {
	service := strings.Split(addr, ":")[0] // ignore :port (we get passed 'host:port')
	_, found := self.Context.GetService(service)
	if !found {
		logrus.Infof("service '%v' not cached; refreshing", service)
		if err := self.Context.RefreshServices(); err != nil {
			return nil, errors.Wrap(err, "error refreshing services")
		}
	}
	return self.Context.Dial(service)
}

type ProxyServiceResolver interface {
	Service(host string) string
}

func NewServiceProxy(p ProxyServiceResolver) (*httputil.ReverseProxy, error) {
	proxy := hostTargetReverseProxy(p)
	director := proxy.Director
	proxy.Director = func(req *http.Request) {
		director(req)
		logrus.Debugf("-> %v", req.URL.String())
		req.Header.Set("X-Proxy", "zrok")
	}
	proxy.ModifyResponse = func(resp *http.Response) error {
		return nil
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logrus.Errorf("error proxying: %v", err)
	}

	return proxy, nil
}

func hostTargetReverseProxy(r ProxyServiceResolver) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		targetSvc := r.Service(req.Host)
		if target, err := url.Parse(fmt.Sprintf("http://%v", targetSvc)); err == nil {
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
