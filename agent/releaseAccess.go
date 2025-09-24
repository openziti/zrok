package agent

import (
	"context"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (a *Agent) ReleaseAccess(frontendToken string) error {
	// first check active accesses
	if acc, found := a.accesses[frontendToken]; found {
		a.rmAccess <- acc
		logrus.Infof("released active access '%v'", acc.frontendToken)
		return nil
	}

	// if not found in active accesses, check failed accesses using frontendToken as failure ID
	if failedEntry, found := a.failedAccesses[frontendToken]; found {
		// mark for removal - retry manager will handle disk persistence
		failedEntry.markedForRemoval = true

		// remove from in-memory failed map immediately so status won't show it
		delete(a.failedAccesses, frontendToken)

		logrus.Infof("removed failed access '%v' (failure id: '%v')", failedEntry.Request.Token, frontendToken)
		return nil
	}

	return errors.Errorf("agent has no access with frontend token or failure id '%v'", frontendToken)
}

func (i *agentGrpcImpl) ReleaseAccess(_ context.Context, req *agentGrpc.ReleaseAccessRequest) (*agentGrpc.ReleaseAccessResponse, error) {
	return nil, i.agent.ReleaseAccess(req.FrontendToken)
}
