package agent

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/agent/proctree"
	"github.com/openziti/zrok/cmd/zrok/subordinate"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/sdk/golang/sdk"
)

func (a *Agent) SharePrivate(req *SharePrivateRequest) (shareToken string, err error) {
	root, err := environment.LoadRoot()
	if err != nil {
		return "", err
	}

	if !root.IsEnabled() {
		return "", errors.New("unable to load environment; did you 'zrok enable'?")
	}

	shrCmd := []string{os.Args[0], "share", "private", "--subordinate", "-b", req.BackendMode}
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
	shr.sub.BootHandler = func(msgType string, msg subordinate.Message) {
		bootErr = shr.bootHandler(msgType, msg)
	}
	shr.sub.MalformedHandler = func(msg subordinate.Message) {
		dl.Error(msg)
	}

	if req.PrivateShareToken != "" {
		shrCmd = append(shrCmd, "--share-token", req.PrivateShareToken)
	}
	if req.Insecure {
		shrCmd = append(shrCmd, "--insecure")
	}
	shr.insecure = req.Insecure

	if !req.Closed {
		shrCmd = append(shrCmd, "--open")
	}
	shr.closed = req.Closed

	for _, grant := range req.AccessGrants {
		shrCmd = append(shrCmd, "--access-grant", grant)
	}
	shr.accessGrants = req.AccessGrants

	shrCmd = append(shrCmd, req.Target)
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
