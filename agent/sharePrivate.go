package agent

import (
	"context"
	"errors"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/agent/proctree"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/sirupsen/logrus"
	"os"
)

func (i *agentGrpcImpl) SharePrivate(_ context.Context, req *agentGrpc.SharePrivateRequest) (*agentGrpc.SharePrivateResponse, error) {
	root, err := environment.LoadRoot()
	if err != nil {
		return nil, err
	}

	if !root.IsEnabled() {
		return nil, errors.New("unable to load environment; did you 'zrok enable'?")
	}

	shrCmd := []string{os.Args[0], "share", "private", "--agent", "-b", req.BackendMode}
	shr := &share{
		shareMode:    sdk.PrivateShareMode,
		backendMode:  sdk.BackendMode(req.BackendMode),
		bootComplete: make(chan struct{}),
		a:            i.a,
	}

	if req.Insecure {
		shrCmd = append(shrCmd, "--insecure")
	}
	shr.insecure = req.Insecure

	if req.Closed {
		shrCmd = append(shrCmd, "--closed")
	}
	shr.closed = req.Closed

	for _, grant := range req.AccessGrants {
		shrCmd = append(shrCmd, "--access-grant", grant)
	}
	shr.accessGrants = req.AccessGrants

	shrCmd = append(shrCmd, req.Target)
	shr.target = req.Target

	logrus.Infof("executing '%v'", shrCmd)

	shr.process, err = proctree.StartChild(shr.tail, shrCmd...)
	if err != nil {
		return nil, err
	}

	go shr.monitor()
	<-shr.bootComplete

	if shr.bootErr == nil {
		i.a.inShares <- shr
		return &agentGrpc.SharePrivateResponse{Token: shr.token}, nil
	}

	return nil, shr.bootErr
}
