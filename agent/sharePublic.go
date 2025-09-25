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
	"github.com/openziti/zrok/sdk/golang/sdk"
)

func (a *Agent) SharePublic(req *SharePublicRequest) (shareToken string, frontendEndpoint []string, err error) {
	root, err := environment.LoadRoot()
	if err != nil {
		return "", nil, err
	}

	if !root.IsEnabled() {
		return "", nil, errors.New("unable to load environment; did you 'zrok enable'?")
	}

	shr := &share{
		shareMode:   sdk.PublicShareMode,
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
	shrCmd := NewSharePublicCommand().
		BackendMode(req.BackendMode).
		BasicAuth(req.BasicAuth).
		NameSelections(req.NameSelections).
		Insecure(req.Insecure).
		OauthProvider(req.OauthProvider).
		OauthEmailDomains(req.OauthEmailDomains).
		OauthRefreshInterval(req.OauthRefreshInterval).
		Open(!req.Closed).
		AccessGrants(req.AccessGrants).
		Target(req.Target).
		Build()

	// set share properties
	shr.basicAuth = req.BasicAuth
	shr.nameSelections = req.NameSelections
	shr.insecure = req.Insecure
	shr.oauthProvider = req.OauthProvider
	shr.oauthEmailAddressPatterns = req.OauthEmailDomains
	shr.closed = req.Closed
	shr.accessGrants = req.AccessGrants
	shr.target = req.Target

	dl.Infof("executing '%v'", shrCmd)

	shr.process, err = proctree.StartChild(shr.sub.Tail, shrCmd...)
	if err != nil {
		return "", nil, err
	}

	<-shr.sub.BootComplete

	if bootErr == nil {
		go shr.monitor()
		a.addShare <- shr
		return shr.token, shr.frontendEndpoints, nil

	} else {
		if err := proctree.WaitChild(shr.process); err != nil {
			dl.Errorf("error joining: %v", err)
		}
		return "", nil, fmt.Errorf("unable to start share: %v", bootErr)
	}
}

func (i *agentGrpcImpl) SharePublic(_ context.Context, req *agentGrpc.SharePublicRequest) (*agentGrpc.SharePublicResponse, error) {
	out := &SharePublicRequest{
		Target:               req.Target,
		BasicAuth:            req.BasicAuth,
		BackendMode:          req.BackendMode,
		Insecure:             req.Insecure,
		OauthProvider:        req.OauthProvider,
		OauthEmailDomains:    req.OauthEmailDomains,
		OauthRefreshInterval: req.OauthRefreshInterval,
		Closed:               req.Closed,
		AccessGrants:         req.AccessGrants,
	}
	for _, nssIn := range req.NameSelections {
		out.NameSelections = append(out.NameSelections, NameSelection{NamespaceToken: nssIn.NamespaceToken, Name: nssIn.Name})
	}
	if shareToken, frontendEndpoints, err := i.agent.SharePublic(out); err == nil {
		return &agentGrpc.SharePublicResponse{Token: shareToken, FrontendEndpoints: frontendEndpoints}, nil
	} else {
		return nil, err
	}
}
