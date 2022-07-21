package http

import (
	"fmt"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"time"
)

func Run(cfg *Config) error {
	options := ziti.ListenOptions{
		ConnectTimeout: 5 * time.Minute,
		MaxConnections: 64,
	}
	zcfg, err := config.NewFromFile(cfg.IdentityPath)
	if err != nil {
		return errors.Wrap(err, "error loading config")
	}
	listener, err := ziti.NewContextWithConfig(zcfg).ListenWithOptions("zrok", &options)
	if err != nil {
		return errors.Wrap(err, "error listening")
	}

	if err := http.Serve(listener, &handler{}); err != nil {
		return err
	}

	return nil
}

type handler struct{}

func (self *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logrus.Infof("handling request from [%v]", r.RemoteAddr)

	r.Host = "localhost:3000"
	r.URL.Host = "localhost:3000"
	r.URL.Scheme = "http"
	r.RequestURI = ""

	rr, err := http.DefaultClient.Do(r)
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
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprint(w, err)
		return
	}

	logrus.Infof("proxied [%d] bytes", n)
}
