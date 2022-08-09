package util

import (
	"net/http"
	"net/http/httputil"
	"sync/atomic"
)

type proxyHandler struct {
	proxy    *httputil.ReverseProxy
	requests int32
}

func NewProxyHandler(proxy *httputil.ReverseProxy) *proxyHandler {
	handler := &proxyHandler{proxy: proxy}

	director := proxy.Director
	proxy.Director = func(req *http.Request) {
		atomic.AddInt32(&handler.requests, 1)
		director(req)
	}

	return handler
}

func (self *proxyHandler) Requests() int32 {
	return atomic.LoadInt32(&self.requests)
}

func (self *proxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	self.proxy.ServeHTTP(w, r)
}
