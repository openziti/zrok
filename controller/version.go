package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/metadata"
)

func versionHandler(_ metadata.VersionParams) middleware.Responder {
	return metadata.NewVersionOK().WithPayload(version)
}
