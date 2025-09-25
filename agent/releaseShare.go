package agent

import (
	"context"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/pkg/errors"
)

func (a *Agent) ReleaseShare(shareToken string) error {
	// first check active shares
	if shr, found := a.shares[shareToken]; found {
		shr.releaseRequested = true
		a.rmShare <- shr
		dl.Infof("released active share '%v'", shr.token)
		return nil
	}

	// then check failed shares
	if a.retryManager.hasFailedShare(shareToken) {
		dl.Infof("released failed share '%v'", shareToken)
		a.retryManager.rmFailedShare(shareToken)
		return nil
	}

	return errors.Errorf("agent has no share with token or failure id '%v'", shareToken)
}

func (i *agentGrpcImpl) ReleaseShare(_ context.Context, req *agentGrpc.ReleaseShareRequest) (*agentGrpc.ReleaseShareResponse, error) {
	return nil, i.agent.ReleaseShare(req.Token)
}
