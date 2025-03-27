package agent

import (
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/zrok/agent/proctree"
	"github.com/openziti/zrok/cmd/zrok/subordinate"
)

type AccessPrivateRequest struct {
	Token           string
	BindAddress     string
	AutoMode        bool
	AutoAddress     string
	AutoStartPort   uint16
	AutoEndPort     uint16
	ResponseHeaders []string
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

	request *AccessPrivateRequest

	process *proctree.Child
	sub     *subordinate.MessageHandler

	agent *Agent
}

func (a *access) monitor() {
	if err := proctree.WaitChild(a.process); err != nil {
		pfxlog.ChannelLogger(a.token).Error(err)
	}
	a.agent.rmAccess <- a
}
