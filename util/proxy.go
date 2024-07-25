package util

import (
	"net/http"
	"net/http/httputil"
	"sync/atomic"
)

type requestsWrapper struct {
	proxy    *httputil.ReverseProxy
	requests int32
}

func NewRequestsWrapper(proxy *httputil.ReverseProxy) *requestsWrapper {
	handler := &requestsWrapper{proxy: proxy}

	director := proxy.Director
	proxy.Director = func(req *http.Request) {
		atomic.AddInt32(&handler.requests, 1)
		director(req)
	}

	return handler
}

func (self *requestsWrapper) Requests() int32 {
	return atomic.LoadInt32(&self.requests)
}

func (self *requestsWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	self.proxy.ServeHTTP(w, r)
}
