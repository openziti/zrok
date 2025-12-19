package agent

import (
	"context"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/agent/agentGrpc"
	"github.com/pkg/errors"
)

func (a *Agent) ReleaseAccess(frontendToken string) error {
	// first check active accesses
	if acc, found := a.accesses[frontendToken]; found {
		acc.releaseRequested = true
		a.rmAccess <- acc
		dl.Infof("released active access '%v'", acc.frontendToken)
		return nil
	}

	// then check failed accesses
	if a.retryManager.hasFailedAccess(frontendToken) {
		dl.Infof("released failed access '%v'", frontendToken)
		a.retryManager.rmFailedAccess(frontendToken)
		return nil
	}

	return errors.Errorf("agent has no access with frontend token or failure id '%v'", frontendToken)
}

func (i *agentGrpcImpl) ReleaseAccess(_ context.Context, req *agentGrpc.ReleaseAccessRequest) (*agentGrpc.ReleaseAccessResponse, error) {
	return nil, i.agent.ReleaseAccess(req.FrontendToken)
}
