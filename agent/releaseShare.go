package agent

import (
	"context"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/agent/proctree"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (i *agentGrpcImpl) ReleaseShare(_ context.Context, req *agentGrpc.ReleaseShareRequest) (*agentGrpc.ReleaseShareResponse, error) {
	if shr, found := i.a.shares[req.Token]; found {
		logrus.Infof("stopping share '%v'", shr.token)

		if err := proctree.StopChild(shr.process); err != nil {
			logrus.Error(err)
		}

		if err := proctree.WaitChild(shr.process); err != nil {
			logrus.Error(err)
		}

		delete(i.a.shares, shr.token)
		logrus.Infof("released share '%v'", shr.token)
	} else {
		return nil, errors.Errorf("agent has no share with token '%v'", req.Token)
	}
	return nil, nil
}
