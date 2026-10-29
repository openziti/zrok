package agent

import (
	"context"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/agent/agentGrpc"
	"github.com/openziti/zrok/v2/build"
)

func (i *agentGrpcImpl) Version(_ context.Context, _ *agentGrpc.VersionRequest) (*agentGrpc.VersionResponse, error) {
	v := build.String()
	dl.Debugf("responding to version inquiry with '%v'", v)
	i.agent.mu.Lock()
	endpoint := i.agent.httpEndpoint
	i.agent.mu.Unlock()
	return &agentGrpc.VersionResponse{
		V:               v,
		ConsoleEndpoint: endpoint,
	}, nil
}
