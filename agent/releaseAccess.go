package agent

import (
	"context"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/agent/proctree"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (i *agentGrpcImpl) ReleaseAccess(_ context.Context, req *agentGrpc.ReleaseAccessRequest) (*agentGrpc.ReleaseAccessReply, error) {
	if acc, found := i.a.accesses[req.FrontendToken]; found {
		logrus.Infof("stopping access '%v'", acc.frontendToken)

		if err := proctree.StopChild(acc.process); err != nil {
			logrus.Error(err)
		}

		if err := proctree.WaitChild(acc.process); err != nil {
			logrus.Error(err)
		}

		delete(i.a.accesses, acc.frontendToken)
		logrus.Infof("released access '%v'", acc.frontendToken)
	} else {
		return nil, errors.Errorf("agent has no access with frontend token '%v'", req.FrontendToken)
	}
	return nil, nil
}
