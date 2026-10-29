package agent

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/openziti/zrok/v2/agent/agentGrpc"
	"github.com/openziti/zrok/v2/environment/env_v0_4"
	"github.com/openziti/zrok/v2/internal/sharepublic"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
)

func (a *Agent) SharePublic(req *SharePublicRequest) (shareToken string, frontendEndpoint []string, err error) {
	return a.sharePublic(context.Background(), req)
}

func (a *Agent) sharePublic(ctx context.Context, req *SharePublicRequest) (string, []string, error) {
	if req == nil {
		return "", nil, fmt.Errorf("missing public share request")
	}
	op, err := a.startRequest(ctx, req)
	if err != nil {
		return "", nil, err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	return op.shr.token, append([]string(nil), op.shr.frontendEndpoints...), nil
}

func (i *agentGrpcImpl) SharePublic(ctx context.Context, req *agentGrpc.SharePublicRequest) (*agentGrpc.SharePublicResponse, error) {
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
	if shareToken, frontendEndpoints, err := i.agent.sharePublic(ctx, out); err == nil {
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
