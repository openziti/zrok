package agent

import (
	"context"
	"errors"
	"fmt"
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

	accCmd := []string{os.Args[0], "access", "private", "--subordinate", "-b", req.BindAddress, req.Token}
	if req.AutoMode {
		accCmd = append(accCmd, "--auto", "--auto-address", req.AutoAddress, "--auto-start-port", fmt.Sprintf("%v", req.AutoStartPort))
		accCmd = append(accCmd, "--auto-end-port", fmt.Sprintf("%v", req.AutoEndPort))
	}
	logrus.Info(accCmd)

	acc := &access{
		token:           req.Token,
		bindAddress:     req.BindAddress,
		autoMode:        req.AutoMode,
		autoAddress:     req.AutoAddress,
		autoStartPort:   uint16(req.AutoStartPort),
		autoEndPort:     uint16(req.AutoEndPort),
		responseHeaders: req.ResponseHeaders,
		bootComplete:    make(chan struct{}),
		agent:           i.agent,
	}

	logrus.Infof("executing '%v'", accCmd)

	acc.process, err = proctree.StartChild(acc.tail, accCmd...)
	if err != nil {
		return nil, err
	}

	<-acc.bootComplete

	if acc.bootErr == nil {
		go acc.monitor()
		i.agent.addAccess <- acc
		return &agentGrpc.AccessPrivateResponse{FrontendToken: acc.frontendToken}, nil
	}
	return nil, acc.bootErr
}
