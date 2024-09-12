package agent

import (
	"context"
	"errors"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/environment"
)

func (i *agentGrpcImpl) PublicShare(_ context.Context, req *agentGrpc.PublicShareRequest) (*agentGrpc.PublicShareReply, error) {
	root, err := environment.LoadRoot()
	if err != nil {
		return nil, err
	}

	if !root.IsEnabled() {
		return nil, errors.New("unable to load environment; did you 'zrok enable'?")
	}

	return &agentGrpc.PublicShareReply{}, nil
}
