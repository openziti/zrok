package http

import (
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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

func (self *handler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	logrus.Infof("handling request from [%v]", req.RemoteAddr)
	if _, err := res.Write([]byte(time.Now().String())); err != nil {
		panic(err)
	}
}
