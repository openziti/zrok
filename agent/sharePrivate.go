package agent

import (
	"context"
	"errors"
	"fmt"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/agent/agentGrpc"
	"github.com/openziti/zrok/v2/agent/proctree"
	"github.com/openziti/zrok/v2/cmd/zrok2/subordinate"
	"github.com/openziti/zrok/v2/environment"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
)

func (a *Agent) SharePrivate(req *SharePrivateRequest) (shareToken string, err error) {
	root, err := environment.LoadRoot()
	if err != nil {
		return "", err
	}

	if !root.IsEnabled() {
		return "", errors.New("unable to load environment; did you 'zrok2 enable'?")
	}

	shr := &share{
		shareMode:   sdk.PrivateShareMode,
		backendMode: sdk.BackendMode(req.BackendMode),
		request:     req,
		sub:         subordinate.NewMessageHandler(),
		agent:       a,
	}
	shr.sub.MessageHandler = func(msg subordinate.Message) {
		dl.Info(msg)
	}
	var bootErr error
	bootHandler := NewShareBootHandler(shr, &bootErr)
	shr.sub.BootHandler = bootHandler.HandleBoot
	shr.sub.MalformedHandler = bootHandler.HandleMalformed

	// build command using CommandBuilder
	shrCmd := NewSharePrivateCommand().
		BackendMode(req.BackendMode).
		ShareToken(req.PrivateShareToken).
		Insecure(req.Insecure).
		Open(!req.Closed).
		AccessGrants(req.AccessGrants).
		Target(req.Target).
		Build()

	// set share properties
	shr.insecure = req.Insecure
	shr.closed = req.Closed
	shr.accessGrants = req.AccessGrants
	shr.target = req.Target

	dl.Infof("executing '%v'", shrCmd)

	shr.process, err = proctree.StartChild(shr.sub.Tail, shrCmd...)
	if err != nil {
		return "", err
	}

	<-shr.sub.BootComplete

	if bootErr == nil {
		go shr.monitor()
		a.addShare <- shr
		return shr.token, nil

	} else {
		if err := proctree.WaitChild(shr.process); err != nil {
			dl.Errorf("error joining: %v", err)
		}
		return "", fmt.Errorf("unable to start share: %v", bootErr)
	}
}

func (i *agentGrpcImpl) SharePrivate(_ context.Context, req *agentGrpc.SharePrivateRequest) (*agentGrpc.SharePrivateResponse, error) {
	if shareToken, err := i.agent.SharePrivate(&SharePrivateRequest{
		Target:            req.Target,
		PrivateShareToken: req.PrivateShareToken,
		BackendMode:       req.BackendMode,
		Insecure:          req.Insecure,
		Closed:            req.Closed,
		AccessGrants:      req.AccessGrants,
	}); err == nil {
		return &agentGrpc.SharePrivateResponse{Token: shareToken}, nil
	} else {
		return nil, err
	}
}
