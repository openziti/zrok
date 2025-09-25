package agent

import (
	"context"
	"errors"
	"fmt"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/agent/proctree"
	"github.com/openziti/zrok/cmd/zrok/subordinate"
	"github.com/openziti/zrok/environment"
)

func (a *Agent) AccessPrivate(req *AccessPrivateRequest) (frontendToken string, err error) {
	root, err := environment.LoadRoot()
	if err != nil {
		return "", err
	}

	if !root.IsEnabled() {
		return "", errors.New("unable to load environment; did you 'zrok enable'?")
	}

	// build command using CommandBuilder
	accCmd := NewAccessPrivateCommand().
		BindAddress(req.BindAddress).
		AutoMode(req.AutoMode, req.AutoAddress, int(req.AutoStartPort), int(req.AutoEndPort)).
		Target(req.ShareToken).
		Build()

	dl.Info(accCmd)

	acc := &access{
		token:           req.ShareToken,
		bindAddress:     req.BindAddress,
		autoMode:        req.AutoMode,
		autoAddress:     req.AutoAddress,
		autoStartPort:   req.AutoStartPort,
		autoEndPort:     req.AutoEndPort,
		responseHeaders: req.ResponseHeaders,
		request:         req,
		sub:             subordinate.NewMessageHandler(),
		agent:           a,
	}
	acc.sub.MessageHandler = func(msg subordinate.Message) {
		dl.Info(msg)
	}
	var bootErr error
	bootHandler := NewAccessBootHandler(acc, &bootErr)
	acc.sub.BootHandler = bootHandler.HandleBoot
	acc.sub.MalformedHandler = bootHandler.HandleMalformed

	dl.Infof("executing '%v'", accCmd)

	acc.process, err = proctree.StartChild(acc.sub.Tail, accCmd...)
	if err != nil {
		return "", err
	}

	<-acc.sub.BootComplete

	if bootErr == nil {
		go acc.monitor()
		a.addAccess <- acc
		return acc.frontendToken, nil

	} else {
		if err := proctree.WaitChild(acc.process); err != nil {
			dl.Errorf("error joining: %v", err)
		}
		return "", fmt.Errorf("unable to start access: %v", bootErr)
	}
}

func (i *agentGrpcImpl) AccessPrivate(_ context.Context, req *agentGrpc.AccessPrivateRequest) (*agentGrpc.AccessPrivateResponse, error) {
	if frontendToken, err := i.agent.AccessPrivate(&AccessPrivateRequest{
		ShareToken:      req.Token,
		BindAddress:     req.BindAddress,
		AutoMode:        req.AutoMode,
		AutoAddress:     req.AutoAddress,
		AutoStartPort:   uint16(req.AutoStartPort),
		AutoEndPort:     uint16(req.AutoEndPort),
		ResponseHeaders: req.ResponseHeaders,
	}); err == nil {
		return &agentGrpc.AccessPrivateResponse{FrontendToken: frontendToken}, nil
	} else {
		return nil, err
	}
}
