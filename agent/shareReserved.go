package agent

import (
	"context"
	"errors"
	"fmt"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/agent/proctree"
	"github.com/openziti/zrok/cmd/zrok/subordinate"
	"github.com/openziti/zrok/environment"
	"github.com/sirupsen/logrus"
	"os"
)

func (i *agentGrpcImpl) ShareReserved(_ context.Context, req *agentGrpc.ShareReservedRequest) (*agentGrpc.ShareReservedResponse, error) {
	root, err := environment.LoadRoot()
	if err != nil {
		return nil, err
	}

	if !root.IsEnabled() {
		return nil, errors.New("unable to load environment; did you 'zrok enable'?")
	}

	shrCmd := []string{os.Args[0], "share", "reserved", "--subordinate"}
	shr := &share{
		reserved: true,
		sub:      subordinate.NewMessageHandler(),
		agent:    i.agent,
	}
	shr.sub.MessageHandler = func(msg subordinate.Message) {
		logrus.Info(msg)
	}
	var bootErr error
	shr.sub.BootHandler = func(msgType string, msg subordinate.Message) {
		bootErr = shr.bootHandler(msgType, msg)
	}
	shr.sub.MalformedHandler = func(msg subordinate.Message) {
		logrus.Error(msg)
	}

	if req.OverrideEndpoint != "" {
		shrCmd = append(shrCmd, "--override-endpoint", req.OverrideEndpoint)
	}

	if req.Insecure {
		shrCmd = append(shrCmd, "--insecure")
	}
	shr.insecure = req.Insecure

	shrCmd = append(shrCmd, req.Token)
	shr.token = req.Token

	shr.process, err = proctree.StartChild(shr.sub.Tail, shrCmd...)
	if err != nil {
		return nil, err
	}

	<-shr.sub.BootComplete

	if bootErr == nil {
		go shr.monitor()
		i.agent.addShare <- shr
		return &agentGrpc.ShareReservedResponse{
			Token:             shr.token,
			BackendMode:       string(shr.backendMode),
			ShareMode:         string(shr.shareMode),
			FrontendEndpoints: shr.frontendEndpoints,
			Target:            shr.target,
		}, nil

	} else {
		if err := proctree.WaitChild(shr.process); err != nil {
			logrus.Errorf("error joining: %v", err)
		}
		return nil, fmt.Errorf("unable to start share: %v", bootErr)
	}
}
