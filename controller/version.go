package controller

import (
	"fmt"
	"regexp"

	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/build"
	"github.com/openziti/zrok/controller/config"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/metadata"
	"github.com/sirupsen/logrus"
)

func versionHandler(_ metadata.VersionParams) middleware.Responder {
	outOfDate := "your local zrok installation is out of date and needs to be upgraded! " +
		"please visit 'https://docs.zrok.io/docs/guides/install/' for the latest release!"
	return metadata.NewVersionOK().WithPayload(rest_model_zrok.Version(outOfDate))
}

type clientVersionCheckHandler struct {
	cfg *config.Config
}

func newClientVersionCheckHandler(cfg *config.Config) *clientVersionCheckHandler {
	return &clientVersionCheckHandler{cfg: cfg}
}

func (h *clientVersionCheckHandler) Handle(params metadata.ClientVersionCheckParams) middleware.Responder {
	logrus.Debugf("client sent version '%v'", params.Body.ClientVersion)

	patterns := h.getVersionPatterns()
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if re.MatchString(params.Body.ClientVersion) {
			logrus.Debugf("client version matched pattern '%v'", pattern)
			return metadata.NewClientVersionCheckOK()
		}
	}

	return metadata.NewClientVersionCheckBadRequest().WithPayload(fmt.Sprintf("client version '%v' does not match any accepted version; please visit 'https://docs.zrok.io/docs/guides/install/' for the latest release!", params.Body.ClientVersion))
}

func (h *clientVersionCheckHandler) getVersionPatterns() []string {
	if h.cfg.Compatibility != nil && len(h.cfg.Compatibility.VersionPatterns) > 0 {
		return h.cfg.Compatibility.VersionPatterns
	}

	// default built-in patterns
	return []string{`^(refs/(heads|tags)/)?` + build.Series}
}

func versionInventoryHandler(params metadata.VersionInventoryParams) middleware.Responder {
	return metadata.NewVersionInventoryOK().WithPayload(&metadata.VersionInventoryOKBody{
		ControllerVersion: build.String(),
	})
}
