package agent

import (
	"context"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (a *Agent) ReleaseShare(shareToken string) error {
	// first check active shares
	if shr, found := a.shares[shareToken]; found {
		a.rmShare <- shr
		logrus.Infof("released active share '%v'", shr.token)
		return nil
	}

	// if not found in active shares, check failed shares using shareToken as failure ID
	if failedEntry, found := a.failedShares[shareToken]; found {
		// mark for removal - retry manager will handle disk persistence
		failedEntry.markedForRemoval = true

		// remove from in-memory failed map immediately so status won't show it
		delete(a.failedShares, shareToken)

		logrus.Infof("removed failed share '%v' (failure id: '%v')", failedEntry.Request.Target, shareToken)
		return nil
	}

	return errors.Errorf("agent has no share with token or failure id '%v'", shareToken)
}

func (i *agentGrpcImpl) ReleaseShare(_ context.Context, req *agentGrpc.ReleaseShareRequest) (*agentGrpc.ReleaseShareResponse, error) {
	return nil, i.agent.ReleaseShare(req.Token)
}
