package controller

import (
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/build"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/metadata"
	"github.com/sirupsen/logrus"
	"regexp"
)

func versionHandler(_ metadata.VersionParams) middleware.Responder {
	outOfDate := "your local zrok installation is out of date and needs to be upgraded! " +
		"please visit 'https://github.com/openziti/zrok/releases' for the latest build!"
	return metadata.NewVersionOK().WithPayload(rest_model_zrok.Version(outOfDate))
}

func clientVersionCheckHandler(params metadata.ClientVersionCheckParams) middleware.Responder {
	logrus.Debugf("client sent version '%v'", params.Body.ClientVersion)
	// allow reported version string to be optionally prefixed with
	// "refs/heads/" or "refs/tags/"
	re := regexp.MustCompile(`^(refs/(heads|tags)/)?` + build.Series)
	if !re.MatchString(params.Body.ClientVersion) {
		return metadata.NewClientVersionCheckBadRequest().WithPayload(fmt.Sprintf("expecting a zrok client version matching '%v' version, received: '%v'; please visit 'https://github.com/openziti/zrok/releases' to make sure you're running the correct client version!", build.Series, params.Body.ClientVersion))
	}
	return metadata.NewClientVersionCheckOK()
}

func versionInventoryHandler(params metadata.VersionInventoryParams) middleware.Responder {
	return metadata.NewVersionInventoryOK().WithPayload(&metadata.VersionInventoryOKBody{
		ControllerVersion: build.String(),
	})
}
