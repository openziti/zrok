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
	data := &rest_model_zrok.Configuration{
		Version:             build.String(),
		InvitesOpen:         cfg.Invites != nil && cfg.Invites.InvitesOpen,
		RequiresInviteToken: cfg.Invites != nil && cfg.Invites.TokenStrategy == "store",
	}
	if cfg.Admin != nil {
		data.TouLink = cfg.Admin.TouLink
	}
	if cfg.Invites != nil {
		data.InviteTokenContact = cfg.Invites.TokenContact
	}
	return metadata.NewConfigurationOK().WithPayload(data)
}
