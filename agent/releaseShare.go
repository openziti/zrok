package agent

import (
	"context"
	"github.com/openziti/zrok/v2/agent/agentGrpc"
)

func (a *Agent) ReleaseShare(token string) error {
	return a.release(context.Background(), token, false)
}
func (i *agentGrpcImpl) ReleaseShare(ctx context.Context, req *agentGrpc.ReleaseShareRequest) (*agentGrpc.ReleaseShareResponse, error) {
	return nil, i.agent.release(ctx, req.Token, false)
}
