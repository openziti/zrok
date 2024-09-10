package agent

import (
	"context"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (i *agentGrpcImpl) ReleaseShare(_ context.Context, req *agentGrpc.ReleaseShareRequest) (*agentGrpc.ReleaseShareReply, error) {
	if shr, found := i.a.shares[req.Token]; found {
		logrus.Infof("stopping share '%v'", shr.shr.Token)
		if err := shr.handler.Stop(); err != nil {
			logrus.Errorf("error stopping share '%v': %v", shr.shr.Token, err)
		}

		root, err := environment.LoadRoot()
		if err != nil {
			return nil, err
		}

		if !root.IsEnabled() {
			return nil, errors.New("unable to load environment; did you 'zrok enable'?")
		}

		if err := sdk.DeleteShare(root, shr.shr); err != nil {
			logrus.Errorf("error releasing share '%v': %v", shr.shr.Token, err)
		}

		delete(i.a.shares, shr.shr.Token)
		logrus.Infof("released share '%v'", shr.shr.Token)
	} else {
		return nil, errors.Errorf("agent has no share with token '%v'", req.Token)
	}
	return nil, nil
}
