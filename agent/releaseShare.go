package agent

import (
	"context"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (a *Agent) ReleaseShare(shareToken string) error {
	if shr, found := a.shares[shareToken]; found {
		a.rmShare <- shr
		logrus.Infof("released share '%v'", shr.token)
	} else {
		errors.Errorf("agent has no share with token '%v'", shareToken)
	}
	return nil
}

func (i *agentGrpcImpl) ReleaseShare(_ context.Context, req *agentGrpc.ReleaseShareRequest) (*agentGrpc.ReleaseShareResponse, error) {
	return nil, i.agent.ReleaseShare(req.Token)
}
