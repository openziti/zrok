package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/metadata"
)

func versionHandler(_ metadata.VersionParams) middleware.Responder {
	outOfDate := "your local zrok installation is out of date and needs to be upgraded! " +
		"please visit 'https://github.com/openziti/zrok/releases' for the latest build!"
	return metadata.NewVersionOK().WithPayload(rest_model_zrok.Version(outOfDate))
}
