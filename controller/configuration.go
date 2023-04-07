package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/build"
	"github.com/openziti/zrok/controller/config"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/metadata"
)

type configurationHandler struct {
	cfg *config.Config
}

func newConfigurationHandler(cfg *config.Config) *configurationHandler {
	return &configurationHandler{
		cfg: cfg,
	}
}

func (ch *configurationHandler) Handle(_ metadata.ConfigurationParams) middleware.Responder {
	tou := ""
	if cfg.Admin != nil {
		tou = cfg.Admin.TouLink
	}
	data := &rest_model_zrok.Configuration{
		Version: build.String(),
		TouLink: tou,
	}
	return metadata.NewConfigurationOK().WithPayload(data)
}
