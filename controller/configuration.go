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
		InvitesOpen:         cfg.Admin != nil && cfg.Admin.InvitesOpen,
		RequiresInviteToken: cfg.Registration != nil && cfg.Admin.InviteTokenStrategy == "store",
	}
	if cfg.Admin != nil {
		data.TouLink = cfg.Admin.TouLink
		data.InviteTokenContact = cfg.Admin.InviteTokenContact
		if cfg.Passwords != nil {
			data.PasswordRequirements = &rest_model_zrok.PasswordRequirements{
				Length:                 int64(cfg.Passwords.Length),
				RequireCapital:         cfg.Passwords.RequireCapital,
				RequireNumeric:         cfg.Passwords.RequireNumeric,
				RequireSpecial:         cfg.Passwords.RequireSpecial,
				ValidSpecialCharacters: cfg.Passwords.ValidSpecialCharacters,
			}
		}
	}
	return metadata.NewConfigurationOK().WithPayload(data)
}
