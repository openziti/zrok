package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/build"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/metadata"
)

func versionHandler(_ metadata.VersionParams) middleware.Responder {
	return metadata.NewVersionOK().WithPayload(rest_model_zrok.Version(build.String()))
}
