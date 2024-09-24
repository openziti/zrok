package agent

import (
	"context"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (i *agentGrpcImpl) ReleaseShare(_ context.Context, req *agentGrpc.ReleaseShareRequest) (*agentGrpc.ReleaseShareResponse, error) {
	if shr, found := i.a.shares[req.Token]; found {
		i.a.outShares <- shr
		logrus.Infof("released share '%v'", shr.token)

	} else {
		return nil, errors.Errorf("agent has no share with token '%v'", req.Token)
	}
	return nil, nil
}
