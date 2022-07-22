package controller

import (
	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/rest_model"
	"github.com/openziti-test-kitchen/zrok/rest_zrok_server"
	"github.com/openziti-test-kitchen/zrok/rest_zrok_server/operations"
	"github.com/openziti-test-kitchen/zrok/rest_zrok_server/operations/metadata"
	"github.com/pkg/errors"
)

func Run(cfg *Config) error {
	swaggerSpec, err := loads.Embedded(rest_zrok_server.SwaggerJSON, rest_zrok_server.FlatSwaggerJSON)
	if err != nil {
		return errors.Wrap(err, "error loading embedded swagger spec")
	}

	api := operations.NewZrokAPI(swaggerSpec)
	api.MetadataGetHandler = metadata.GetHandlerFunc(func(params metadata.GetParams) middleware.Responder {
		return metadata.NewGetOK().WithPayload(&rest_model.Version{Version: "oh, wow!"})
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
