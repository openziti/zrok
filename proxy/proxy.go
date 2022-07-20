package proxy

import (
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/config"
	"github.com/pkg/errors"
	"net/http"
	"time"
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
	_, err := w.Write([]byte(time.Now().String()))
	if err != nil {
		panic(err)
	}
}
