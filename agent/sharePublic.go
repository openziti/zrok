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

func (i *agentGrpcImpl) SharePublic(_ context.Context, req *agentGrpc.SharePublicRequest) (*agentGrpc.SharePublicResponse, error) {
	root, err := environment.LoadRoot()
	if err != nil {
		return nil, err
	}

	if !root.IsEnabled() {
		return nil, errors.New("unable to load environment; did you 'zrok enable'?")
	}

	shrCmd := []string{os.Args[0], "share", "public", "--subordinate", "-b", req.BackendMode}
	shr := &share{
		shareMode:    sdk.PublicShareMode,
		backendMode:  sdk.BackendMode(req.BackendMode),
		bootComplete: make(chan struct{}),
		agent:        i.agent,
	}

	for _, basicAuth := range req.BasicAuth {
		shrCmd = append(shrCmd, "--basic-auth", basicAuth)
	}
	shr.basicAuth = req.BasicAuth

	for _, frontendSelection := range req.FrontendSelection {
		shrCmd = append(shrCmd, "--frontend", frontendSelection)
	}
	shr.frontendSelection = req.FrontendSelection

	if req.Insecure {
		shrCmd = append(shrCmd, "--insecure")
	}
	shr.insecure = req.Insecure

	if req.OauthProvider != "" {
		shrCmd = append(shrCmd, "--oauth-provider", req.OauthProvider)
	}
	shr.oauthProvider = req.OauthProvider

	for _, pattern := range req.OauthEmailAddressPatterns {
		shrCmd = append(shrCmd, "--oauth-email-address-patterns", pattern)
	}
	shr.oauthEmailAddressPatterns = req.OauthEmailAddressPatterns

	if req.OauthCheckInterval != "" {
		shrCmd = append(shrCmd, "--oauth-check-interval", req.OauthCheckInterval)
	}

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
		i.agent.addShare <- shr
		return &agentGrpc.SharePublicResponse{
			Token:             shr.token,
			FrontendEndpoints: shr.frontendEndpoints,
		}, nil
	}

	return nil, shr.bootErr
}