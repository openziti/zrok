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
	if h.cfg.Compatibility != nil && h.cfg.Compatibility.LogRequests {
		logrus.Infof("client at '%v' sent version '%v'", params.HTTPRequest.RemoteAddr, params.Body.ClientVersion)
	}

	patterns := h.getCompiledPatterns()
	for i, re := range patterns {
		if re.MatchString(params.Body.ClientVersion) {
			logrus.Debugf("client version matched pattern %d", i)
			return metadata.NewClientVersionCheckOK()
		}
	}

	return metadata.NewClientVersionCheckBadRequest().WithPayload(fmt.Sprintf("client version '%v' does not match any accepted patterns; please visit 'https://docs.zrok.io/docs/guides/install/' for the latest release!", params.Body.ClientVersion))
}

func (h *clientVersionCheckHandler) getCompiledPatterns() []*regexp.Regexp {
	if h.cfg.Compatibility != nil && len(h.cfg.Compatibility.GetCompiledPatterns()) > 0 {
		return h.cfg.Compatibility.GetCompiledPatterns()
	}

	// fallback to built-in patterns (this should not happen in normal operation)
	logrus.Errorf("missing compatibility patterns; defaulting to last-resort patterns")
	defaultPatterns := []*regexp.Regexp{
		regexp.MustCompile(`^(refs/(heads|tags)/)?` + build.Series),
	}
	return defaultPatterns
}

func versionInventoryHandler(params metadata.VersionInventoryParams) middleware.Responder {
	return metadata.NewVersionInventoryOK().WithPayload(&metadata.VersionInventoryOKBody{
		ControllerVersion: build.String(),
	})
}
