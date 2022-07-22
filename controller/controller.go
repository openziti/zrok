package controller

import (
	"github.com/go-openapi/loads"
	"github.com/openziti-test-kitchen/zrok/rest_zrok_server"
	"github.com/openziti-test-kitchen/zrok/rest_zrok_server/operations"
	"github.com/pkg/errors"
)

func Run(cfg *Config) error {
	swaggerSpec, err := loads.Embedded(rest_zrok_server.SwaggerJSON, rest_zrok_server.FlatSwaggerJSON)
	if err != nil {
		return errors.Wrap(err, "error loading embedded swagger spec")
	}

	api := operations.NewZrokAPI(swaggerSpec)
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
