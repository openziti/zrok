package agent

import (
	"context"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/agent/agentGrpc"
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

	// then check failed public shares
	if a.retryManager.hasFailedPublicShare(shareToken) {
		dl.Infof("released failed public share '%v'", shareToken)
		a.retryManager.rmFailedPublicShare(shareToken)
		return nil
	}

	// then check failed private shares
	if a.retryManager.hasFailedPrivateShare(shareToken) {
		dl.Infof("released failed private share '%v'", shareToken)
		a.retryManager.rmFailedPrivateShare(shareToken)
		return nil
	}

	return errors.Errorf("agent has no share with token or failure id '%v'", shareToken)
}

func (i *agentGrpcImpl) ReleaseShare(_ context.Context, req *agentGrpc.ReleaseShareRequest) (*agentGrpc.ReleaseShareResponse, error) {
	return nil, i.agent.ReleaseShare(req.Token)
}
