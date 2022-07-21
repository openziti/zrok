package proxy

import (
	"fmt"
	"github.com/openziti-test-kitchen/zrok/util"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
)

func Run(cfg *Config) error {
	zCfg, err := config.NewFromFile(cfg.IdentityPath)
	if err != nil {
		return errors.Wrap(err, "error loading config")
	}
	zCtx := ziti.NewContextWithConfig(zCfg)
	handler := &handler{
		zCfg: zCfg,
		zCtx: zCtx,
	}

	return http.ListenAndServe(cfg.Address, handler)
}

type handler struct {
	zCfg *config.Config
	zCtx ziti.Context
}

func (self *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logrus.Warnf("handling request from [%v]", r.RemoteAddr)

	r.Host = "zrok"
	r.URL.Host = "zrok"
	r.URL.Scheme = "http"
	r.RequestURI = ""
	logrus.Info(util.DumpHeaders(r.Header, true))

	logrus.Infof("forwarding to: %v [%v]", r.Method, r.URL)
	zDialCtx := util.ZitiDialContext{Context: self.zCtx}
	zTransport := http.DefaultTransport.(*http.Transport).Clone()
	zTransport.DialContext = zDialCtx.Dial
	zClient := &http.Client{Transport: zTransport}
	rr, err := zClient.Do(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprint(w, err)
		return
	}
	w.WriteHeader(rr.StatusCode)
	logrus.Infof("response: %v", rr.Status)

	// forward headers
	for k, v := range rr.Header {
		for _, vi := range v {
			w.Header().Add(k, vi)
		}
	}
	logrus.Info(util.DumpHeaders(w.Header(), false))

	// copy body
	n, err := io.Copy(w, rr.Body)
	if err != nil {
		panic(err)
	}

	logrus.Infof("proxied [%d] bytes", n)
}
