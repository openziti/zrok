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
	logrus.Infof("handling request from [%v]", r.RemoteAddr)

	zDialCtx := util.ZitiDialContext{Context: self.zCtx}
	zTransport := http.DefaultTransport.(*http.Transport).Clone()
	zTransport.DialContext = zDialCtx.Dial
	client := &http.Client{Transport: zTransport}
	r.Host = "zrok"
	r.URL.Host = "zrok"
	r.URL.Scheme = "http"
	r.RequestURI = ""
	logrus.Warnf("request: %v", r)

	rr, err := client.Do(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprint(w, err)
		return
	}

	for k, v := range rr.Header {
		w.Header().Add(k, v[0])
	}

	n, err := io.Copy(w, rr.Body)
	if err != nil {
		panic(err)
	}

	logrus.Infof("proxied [%d] bytes", n)
}
