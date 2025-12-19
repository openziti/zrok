package agent

import (
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/agent/proctree"
	"github.com/openziti/zrok/v2/cmd/zrok/subordinate"
)

type AccessPrivateRequest struct {
	ShareToken      string   `json:"share_token"`
	BindAddress     string   `json:"bind_address"`
	AutoMode        bool     `json:"auto_mode"`
	AutoAddress     string   `json:"auto_address"`
	AutoStartPort   uint16   `json:"auto_start_port"`
	AutoEndPort     uint16   `json:"auto_end_port"`
	ResponseHeaders []string `json:"response_headers"`
}

type access struct {
	frontendToken   string
	token           string
	bindAddress     string
	autoMode        bool
	autoAddress     string
	autoStartPort   uint16
	autoEndPort     uint16
	responseHeaders []string

	request          *AccessPrivateRequest
	releaseRequested bool
	processExited    bool
	lastError        error

	process *proctree.Child
	sub     *subordinate.MessageHandler

	agent *Agent
}

func (a *access) monitor() {
	if err := proctree.WaitChild(a.process); err != nil {
		dl.ChannelLog(a.token).Error(err)
		a.lastError = err
	}
	a.processExited = true
	a.agent.rmAccess <- a
}
