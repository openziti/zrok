package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/zrok/agent/proctree"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

type share struct {
	token                     string
	frontendEndpoints         []string
	target                    string
	basicAuth                 []string
	frontendSelection         []string
	shareMode                 sdk.ShareMode
	backendMode               sdk.BackendMode
	reserved                  bool
	insecure                  bool
	oauthProvider             string
	oauthEmailAddressPatterns []string
	oauthCheckInterval        time.Duration
	closed                    bool
	accessGrants              []string

	process      *proctree.Child
	readBuffer   bytes.Buffer
	booted       bool
	bootComplete chan struct{}
	bootErr      error

	agent *Agent
}

func (s *share) monitor() {
	if err := proctree.WaitChild(s.process); err != nil {
		pfxlog.ChannelLogger(s.token).Error(err)
	}
	s.agent.rmShare <- s
}

func (s *share) tail(data []byte) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("recovering: %v", r)
		}
	}()
	s.readBuffer.Write(data)
	if line, err := s.readBuffer.ReadString('\n'); err == nil {
		line = strings.Trim(line, "\n")
		if !s.booted {
			if strings.HasPrefix(line, "{") {
				in := make(map[string]interface{})
				if err := json.Unmarshal([]byte(line), &in); err == nil {
					if v, found := in["message"]; found {
						if str, ok := v.(string); ok {
							if str == "boot" {
								if v, found := in["token"]; found {
									if str, ok := v.(string); ok {
										s.token = str
									}
								}
								if v, found := in["backend_mode"]; found {
									if str, ok := v.(string); ok {
										s.backendMode = sdk.BackendMode(str)
									}
								}
								if v, found := in["share_mode"]; found {
									if str, ok := v.(string); ok {
										s.shareMode = sdk.ShareMode(str)
									}
								}
								if v, found := in["frontend_endpoints"]; found {
									if vArr, ok := v.([]interface{}); ok {
										for _, v := range vArr {
											if str, ok := v.(string); ok {
												s.frontendEndpoints = append(s.frontendEndpoints, str)
											}
										}
									}
								}
								if v, found := in["target"]; found {
									if str, ok := v.(string); ok {
										s.target = str
									}
								}
								s.booted = true
							} else {
								s.bootErr = errors.New(line)
							}
						} else {
							s.bootErr = errors.New(line)
						}
					} else {
						s.bootErr = errors.New(line)
					}
				} else {
					s.bootErr = errors.New(line)
				}
				close(s.bootComplete)
			} else {
				logrus.Warn(line)
			}
		} else {
			if strings.HasPrefix(line, "{") {
				in := make(map[string]interface{})
				if err := json.Unmarshal([]byte(line), &in); err == nil {
					pfxlog.ChannelLogger(s.token).Info(in)
				}
			} else {
				pfxlog.ChannelLogger(s.token).Info(strings.Trim(line, "\n"))
			}
		}
	} else {
		s.readBuffer.WriteString(line)
	}
}
