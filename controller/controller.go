package controller

import (
	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/controller/store"
	"github.com/openziti-test-kitchen/zrok/rest_model"
	"github.com/openziti-test-kitchen/zrok/rest_zrok_server"
	"github.com/openziti-test-kitchen/zrok/rest_zrok_server/operations"
	"github.com/openziti-test-kitchen/zrok/rest_zrok_server/operations/metadata"
	"github.com/pkg/errors"
)

var str *store.Store

func Run(cfg *Config) error {
	if v, err := store.Open(cfg.Store); err == nil {
		str = v
	} else {
		return errors.Wrap(err, "error opening store")
	}

	swaggerSpec, err := loads.Embedded(rest_zrok_server.SwaggerJSON, rest_zrok_server.FlatSwaggerJSON)
	if err != nil {
		return errors.Wrap(err, "error loading embedded swagger spec")
	}

	api := operations.NewZrokAPI(swaggerSpec)
	api.MetadataVersionHandler = metadata.VersionHandlerFunc(func(_ metadata.VersionParams) middleware.Responder {
		return metadata.NewGetOK().WithPayload(&rest_model.Version{Version: "v0.0.0; sk3tch"})
	})

	server := rest_zrok_server.NewServer(api)
	defer func() { _ = server.Shutdown() }()
	server.Host = cfg.Host
	server.Port = cfg.Port
	server.ConfigureAPI()
	if err := server.Serve(); err != nil {
		return errors.Wrap(err, "api server error")
	}
	return nil
}
