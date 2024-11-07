package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/zrok/agent/proctree"
	"github.com/sirupsen/logrus"
	"strings"
)

type access struct {
	frontendToken   string
	token           string
	bindAddress     string
	autoMode        bool
	autoAddress     string
	autoStartPort   uint16
	autoEndPort     uint16
	responseHeaders []string

	process      *proctree.Child
	readBuffer   bytes.Buffer
	booted       bool
	bootComplete chan struct{}
	bootErr      error

	agent *Agent
}

func (a *access) monitor() {
	if err := proctree.WaitChild(a.process); err != nil {
		pfxlog.ChannelLogger(a.token).Error(err)
	}
	a.agent.rmAccess <- a
}

func (a *access) tail(data []byte) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("recovering: %v", r)
		}
	}()
	a.readBuffer.Write(data)
	if line, err := a.readBuffer.ReadString('\n'); err == nil {
		line = strings.Trim(line, "\n")
		if !a.booted {
			in := make(map[string]interface{})
			if err := json.Unmarshal([]byte(line), &in); err == nil {
				if v, found := in["frontend_token"]; found {
					if str, ok := v.(string); ok {
						a.frontendToken = str
					}
				}
				if v, found := in["bind_address"]; found {
					if str, ok := v.(string); ok {
						a.bindAddress = str
					}
				}
				a.booted = true
			} else {
				a.bootErr = errors.New(line)
			}
			close(a.bootComplete)

		} else {
			if strings.HasPrefix(line, "{") {
				in := make(map[string]interface{})
				if err := json.Unmarshal([]byte(line), &in); err == nil {
					pfxlog.ChannelLogger(a.token).Info(in)
				}
			} else {
				pfxlog.ChannelLogger(a.token).Info(strings.Trim(line, "\n"))
			}
		}
	} else {
		a.readBuffer.WriteString(line)
	}
}
