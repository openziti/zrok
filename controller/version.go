package controller

import (
	"fmt"
	"regexp"

	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/build"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/metadata"
	"github.com/sirupsen/logrus"
)

func versionHandler(_ metadata.VersionParams) middleware.Responder {
	outOfDate := "your local zrok installation is out of date and needs to be upgraded! " +
		"please visit 'https://docs.zrok.io/docs/guides/install/' for the latest release!"
	return metadata.NewVersionOK().WithPayload(rest_model_zrok.Version(outOfDate))
}

func clientVersionCheckHandler(params metadata.ClientVersionCheckParams) middleware.Responder {
	logrus.Debugf("client sent version '%v'", params.Body.ClientVersion)

	// allow reported version string to be optionally prefixed with
	// "refs/heads/" or "refs/tags/"
	currentVersion := regexp.MustCompile(`^(refs/(heads|tags)/)?` + build.Series)
	if currentVersion.MatchString(params.Body.ClientVersion) {
		logrus.Debugf("client version matched current version stream '%v'", build.Series)
		return metadata.NewClientVersionCheckOK()
	}

	previousVersion := regexp.MustCompile(`^v1.0`)
	if previousVersion.MatchString(params.Body.ClientVersion) {
		logrus.Debug("client version matched previous version stream 'v1.0'")
		return metadata.NewClientVersionCheckOK()
	}

	return metadata.NewClientVersionCheckBadRequest().WithPayload(fmt.Sprintf("expecting a zrok client version matching '%v' version (or previous version 'v1.0'), received: '%v'; please visit 'https://docs.zrok.io/docs/guides/install/' for the latest release!", build.Series, params.Body.ClientVersion))
}

func versionInventoryHandler(params metadata.VersionInventoryParams) middleware.Responder {
	return metadata.NewVersionInventoryOK().WithPayload(&metadata.VersionInventoryOKBody{
		ControllerVersion: build.String(),
	})
}
