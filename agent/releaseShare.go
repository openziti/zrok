package agent

import (
	"context"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/sirupsen/logrus"
)

func (i *agentGrpcImpl) ReleaseShare(_ context.Context, req *agentGrpc.ReleaseShareRequest) (*agentGrpc.ReleaseShareReply, error) {
	logrus.Infof("releasing '%v'", req.Token)
	return nil, nil
}
