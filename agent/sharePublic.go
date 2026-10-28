package agent

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/agent/agentGrpc"
	"github.com/openziti/zrok/v2/agent/proctree"
	"github.com/openziti/zrok/v2/cmd/zrok2/subordinate"
	"github.com/openziti/zrok/v2/environment/env_v0_4"
	"github.com/openziti/zrok/v2/internal/sharepublic"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
)

func (a *Agent) SharePublic(req *SharePublicRequest) (shareToken string, frontendEndpoint []string, err error) {
	prepared, err := a.preparePublicRequest(req)
	if err != nil {
		return "", nil, err
	}
	original := req
	normalized := *req
	normalized.Target = prepared.Target
	normalized.BasicAuth = prepared.BasicAuth
	normalized.AccessGrants = prepared.AccessGrants
	normalized.OauthEmailDomains = prepared.OauthEmailAddressPatterns
	normalized.OauthRefreshInterval = prepared.OauthRefreshInterval.String()
	normalized.NameSelections = nil
	for _, selection := range prepared.NameSelections {
		normalized.NameSelections = append(normalized.NameSelections, NameSelection{NamespaceToken: selection.NamespaceToken, Name: selection.Name})
	}
	req = &normalized

	shr := &share{
		shareMode:   sdk.PublicShareMode,
		backendMode: sdk.BackendMode(req.BackendMode),
		request:     original,
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
		if nssIn == nil {
			return nil, fmt.Errorf("nil name selection")
		}
		out.NameSelections = append(out.NameSelections, NameSelection{NamespaceToken: nssIn.NamespaceToken, Name: nssIn.Name})
	}
	if shareToken, frontendEndpoints, err := i.agent.SharePublic(out); err == nil {
		return &agentGrpc.SharePublicResponse{Token: shareToken, FrontendEndpoints: frontendEndpoints}, nil
	} else {
		return nil, err
	}
}

// preparePublicRequest adapts agent transport settings without changing saved request intent or credentials.
func (a *Agent) preparePublicRequest(req *SharePublicRequest) (*sdk.ShareRequest, error) {
	if req == nil {
		return nil, fmt.Errorf("missing public share request")
	}
	interval := 3 * time.Hour
	if req.OauthRefreshInterval != "" {
		var err error
		interval, err = time.ParseDuration(req.OauthRefreshInterval)
		if err != nil {
			return nil, fmt.Errorf("invalid oauth refresh interval '%v': %w", req.OauthRefreshInterval, err)
		}
	}
	target, err := sharepublic.Target(req.BackendMode, []string{req.Target})
	if err != nil {
		return nil, err
	}
	if req.BackendMode != "proxy" {
		target, err = filepath.Abs(target)
		if err != nil {
			return nil, err
		}
	}
	in := sdk.ShareRequest{
		BackendMode: sdk.BackendMode(req.BackendMode), Target: target,
		BasicAuth: req.BasicAuth, AccessGrants: req.AccessGrants,
		OauthProvider: req.OauthProvider, OauthEmailAddressPatterns: req.OauthEmailDomains,
		OauthRefreshInterval: interval, PermissionMode: sdk.OpenPermissionMode,
	}
	if req.Closed {
		in.PermissionMode = sdk.ClosedPermissionMode
	}
	for _, selection := range req.NameSelections {
		in.NameSelections = append(in.NameSelections, sdk.NameSelection{NamespaceToken: selection.NamespaceToken, Name: selection.Name})
	}
	var defaultNamespace string
	if len(in.NameSelections) == 0 {
		defaultNamespace, err = env_v0_4.FreshDefaultNamespace(a.root)
		if err != nil {
			return nil, err
		}
	}
	return sharepublic.Prepare(in, defaultNamespace)
}
