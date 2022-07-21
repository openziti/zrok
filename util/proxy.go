package util

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func NewProxy(target string) (*httputil.ReverseProxy, error) {
	targetURL, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
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
	}

	return proxy, nil
}

type proxyHandler struct {
	proxy *httputil.ReverseProxy
}

func NewProxyHandler(proxy *httputil.ReverseProxy) *proxyHandler {
	return &proxyHandler{proxy}
}

func (self *proxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logrus.Infof("proxying from: %v", r.RequestURI)
	self.proxy.ServeHTTP(w, r)
}
