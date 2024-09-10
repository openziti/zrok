package agent

import (
	"context"
	"errors"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/endpoints/proxy"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/sirupsen/logrus"
	"time"
)

func (i *agentGrpcImpl) PublicShare(_ context.Context, req *agentGrpc.PublicShareRequest) (*agentGrpc.PublicShareReply, error) {
	root, err := environment.LoadRoot()
	if err != nil {
		return nil, err
	}

	if !root.IsEnabled() {
		return nil, errors.New("unable to load environment; did you 'zrok enable'?")
	}

	zif, err := root.ZitiIdentityNamed(root.EnvironmentIdentityName())
	if err != nil {
		return nil, err
	}

	shrReq := &sdk.ShareRequest{
		BackendMode: sdk.BackendMode(req.BackendMode),
		ShareMode:   sdk.PublicShareMode,
		Frontends:   req.FrontendSelection,
		BasicAuth:   req.BasicAuth,
		Target:      req.Target,
	}
	if req.Closed {
		shrReq.PermissionMode = sdk.ClosedPermissionMode
		shrReq.AccessGrants = req.AccessGrants
	}
	if req.OauthProvider != "" {
		shrReq.OauthProvider = req.OauthProvider
		shrReq.OauthEmailAddressPatterns = req.OauthEmailAddressPatterns
		checkInterval, err := time.ParseDuration(req.GetOauthCheckInterval())
		if err != nil {
			return nil, err
		}
		shrReq.OauthAuthorizationCheckInterval = checkInterval
	}
	shr, err := sdk.CreateShare(root, shrReq)
	if err != nil {
		return nil, err
	}

	switch req.BackendMode {
	case "proxy":
		cfg := &proxy.BackendConfig{
			IdentityPath:    zif,
			EndpointAddress: req.Target,
			ShrToken:        shr.Token,
			Insecure:        req.Insecure,
		}

		be, err := proxy.NewBackend(cfg)
		if err != nil {
			return nil, err
		}

		agentShr := &share{
			shr:                       shr,
			target:                    req.Target,
			basicAuth:                 req.BasicAuth,
			frontendSelection:         shr.FrontendEndpoints,
			shareMode:                 sdk.PublicShareMode,
			backendMode:               sdk.BackendMode(req.BackendMode),
			insecure:                  req.Insecure,
			oauthProvider:             req.OauthProvider,
			oauthEmailAddressPatterns: req.OauthEmailAddressPatterns,
			oauthCheckInterval:        shrReq.OauthAuthorizationCheckInterval,
			closed:                    req.Closed,
			accessGrants:              req.AccessGrants,
			handler:                   be,
		}

		i.a.shares[shr.Token] = agentShr
		go func() {
			if err := agentShr.handler.Run(); err != nil {
				logrus.Errorf("error running proxy backend: %v", err)
			}
		}()

	case "web":
		cfg := &proxy.CaddyWebBackendConfig{
			IdentityPath: zif,
			WebRoot:      req.Target,
			ShrToken:     shr.Token,
		}

		be, err := proxy.NewCaddyWebBackend(cfg)
		if err != nil {
			return nil, err
		}

		agentShr := &share{
			shr:                       shr,
			target:                    req.Target,
			basicAuth:                 req.BasicAuth,
			frontendSelection:         shr.FrontendEndpoints,
			shareMode:                 sdk.PublicShareMode,
			backendMode:               sdk.BackendMode(req.BackendMode),
			insecure:                  req.Insecure,
			oauthProvider:             req.OauthProvider,
			oauthEmailAddressPatterns: req.OauthEmailAddressPatterns,
			oauthCheckInterval:        shrReq.OauthAuthorizationCheckInterval,
			closed:                    req.Closed,
			accessGrants:              req.AccessGrants,
			handler:                   be,
		}

		i.a.shares[shr.Token] = agentShr
		go func() {
			if err := agentShr.handler.Run(); err != nil {
				logrus.Errorf("error running web backend: %v", err)
			}
		}()
	}

	return &agentGrpc.PublicShareReply{Token: shr.Token}, nil
}
