package agent

import (
	"context"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (i *agentGrpcImpl) ReleaseAccess(_ context.Context, req *agentGrpc.ReleaseAccessRequest) (*agentGrpc.ReleaseAccessResponse, error) {
	if acc, found := i.agent.accesses[req.FrontendToken]; found {
		i.agent.rmAccess <- acc
		logrus.Infof("released access '%v'", acc.frontendToken)

	} else {
		return nil, errors.Errorf("agent has no access with frontend token '%v'", req.FrontendToken)
	}
	return nil, nil
}
