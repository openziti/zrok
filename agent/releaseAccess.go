package agent

import (
	"context"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (a *Agent) ReleaseAccess(frontendToken string) error {
	if acc, found := a.accesses[frontendToken]; found {
		a.rmAccess <- acc
		logrus.Infof("released access '%v'", acc.frontendToken)
	} else {
		return errors.Errorf("agent has no access with frontend token '%v'", frontendToken)
	}
	return nil
}

func (i *agentGrpcImpl) ReleaseAccess(_ context.Context, req *agentGrpc.ReleaseAccessRequest) (*agentGrpc.ReleaseAccessResponse, error) {
	return nil, i.agent.ReleaseAccess(req.FrontendToken)
}
