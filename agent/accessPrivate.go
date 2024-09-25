package agent

import (
	"context"
	"errors"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/agent/proctree"
	"github.com/openziti/zrok/environment"
	"github.com/sirupsen/logrus"
	"os"
)

func (i *agentGrpcImpl) AccessPrivate(_ context.Context, req *agentGrpc.AccessPrivateRequest) (*agentGrpc.AccessPrivateResponse, error) {
	root, err := environment.LoadRoot()
	if err != nil {
		return nil, err
	}

	if !root.IsEnabled() {
		return nil, errors.New("unable to load environment; did you 'zrok enable'?")
	}

	accCmd := []string{os.Args[0], "access", "private", "--agent", "-b", req.BindAddress, req.Token}
	acc := &access{
		token:           req.Token,
		bindAddress:     req.BindAddress,
		responseHeaders: req.ResponseHeaders,
		bootComplete:    make(chan struct{}),
		agent:           i.agent,
	}

	logrus.Infof("executing '%v'", accCmd)

	acc.process, err = proctree.StartChild(acc.tail, accCmd...)
	if err != nil {
		return nil, err
	}

	go acc.monitor()
	<-acc.bootComplete

	if acc.bootErr == nil {
		i.agent.addAccess <- acc
		return &agentGrpc.AccessPrivateResponse{FrontendToken: acc.frontendToken}, nil
	}

	return nil, acc.bootErr
}
