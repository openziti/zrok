package agent

import (
	"context"
	"errors"

	"github.com/openziti/zrok/v2/agent/agentGrpc"
)

func (a *Agent) SharePrivate(req *SharePrivateRequest) (string, error) {
	return a.sharePrivate(context.Background(), req)
}
func (a *Agent) sharePrivate(ctx context.Context, req *SharePrivateRequest) (string, error) {
	if req == nil {
		return "", errors.New("missing private share request")
	}
	op, err := a.startRequest(ctx, req)
	if err != nil {
		return "", err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	return op.shr.token, nil
}

func (i *agentGrpcImpl) SharePrivate(ctx context.Context, req *agentGrpc.SharePrivateRequest) (*agentGrpc.SharePrivateResponse, error) {
	if shareToken, err := i.agent.sharePrivate(ctx, &SharePrivateRequest{
		Target:            req.Target,
		PrivateShareToken: req.PrivateShareToken,
		BackendMode:       req.BackendMode,
		Insecure:          req.Insecure,
		Closed:            req.Closed,
		AccessGrants:      req.AccessGrants,
	}); err == nil {
		return &agentGrpc.SharePrivateResponse{Token: shareToken}, nil
	} else {
		return nil, err
	}
}
