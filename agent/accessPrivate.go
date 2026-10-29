package agent

import (
	"context"
	"errors"

	"github.com/openziti/zrok/v2/agent/agentGrpc"
)

func (a *Agent) AccessPrivate(req *AccessPrivateRequest) (string, error) {
	return a.accessPrivate(context.Background(), req)
}
func (a *Agent) accessPrivate(ctx context.Context, req *AccessPrivateRequest) (string, error) {
	if req == nil {
		return "", errors.New("missing private access request")
	}
	op, err := a.startRequest(ctx, req)
	if err != nil {
		return "", err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	return op.acc.frontendToken, nil
}

func (i *agentGrpcImpl) AccessPrivate(ctx context.Context, req *agentGrpc.AccessPrivateRequest) (*agentGrpc.AccessPrivateResponse, error) {
	if frontendToken, err := i.agent.accessPrivate(ctx, &AccessPrivateRequest{
		ShareToken:      req.Token,
		BindAddress:     req.BindAddress,
		AutoMode:        req.AutoMode,
		AutoAddress:     req.AutoAddress,
		AutoStartPort:   uint16(req.AutoStartPort),
		AutoEndPort:     uint16(req.AutoEndPort),
		ResponseHeaders: req.ResponseHeaders,
	}); err == nil {
		return &agentGrpc.AccessPrivateResponse{FrontendToken: frontendToken}, nil
	} else {
		return nil, err
	}
}
