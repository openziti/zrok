package agent

import (
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/agent/proctree"
)

type access struct {
	token string

	bindAddress     string
	responseHeaders []string

	process *proctree.Child
}

type agentGrpcImpl struct {
	agentGrpc.UnimplementedAgentServer
	a *Agent
}
