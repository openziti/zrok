package agent

import (
	"context"
	"github.com/openziti/zrok/v2/agent/agentGrpc"
)

func (a *Agent) ReleaseAccess(token string) error {
	return a.release(context.Background(), token, true)
}
func (i *agentGrpcImpl) ReleaseAccess(ctx context.Context, req *agentGrpc.ReleaseAccessRequest) (*agentGrpc.ReleaseAccessResponse, error) {
	return nil, i.agent.release(ctx, req.FrontendToken, true)
}
