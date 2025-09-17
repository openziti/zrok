package agent

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/agent/proctree"
	"github.com/openziti/zrok/cmd/zrok/subordinate"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/sirupsen/logrus"
)

func (a *Agent) SharePublic(req *SharePublicRequest) (shareToken string, frontendEndpoint []string, err error) {
	root, err := environment.LoadRoot()
	if err != nil {
		return "", nil, err
	}

	if !root.IsEnabled() {
		return "", nil, errors.New("unable to load environment; did you 'zrok enable'?")
	}

	shrCmd := []string{os.Args[0], "share", "public", "--subordinate", "-b", req.BackendMode}
	shr := &share{
		shareMode:   sdk.PublicShareMode,
		backendMode: sdk.BackendMode(req.BackendMode),
		request:     req,
		sub:         subordinate.NewMessageHandler(),
		agent:       a,
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

	for _, basicAuth := range req.BasicAuth {
		shrCmd = append(shrCmd, "--basic-auth", basicAuth)
	}
	shr.basicAuth = req.BasicAuth

	for _, nss := range req.NamespaceSelections {
		nssStr := nss.NamespaceToken
		if nss.Name != "" {
			nssStr += ":" + nss.Name
		}
		shrCmd = append(shrCmd, "--namespace-selection", nssStr)
	}
	shr.namespaceSelections = req.NamespaceSelections

	if req.Insecure {
		shrCmd = append(shrCmd, "--insecure")
	}
	shr.insecure = req.Insecure

	if req.OauthProvider != "" {
		shrCmd = append(shrCmd, "--oauth-provider", req.OauthProvider)
	}
	shr.oauthProvider = req.OauthProvider

	for _, pattern := range req.OauthEmailDomains {
		shrCmd = append(shrCmd, "--oauth-email-domain", pattern)
	}
	shr.oauthEmailAddressPatterns = req.OauthEmailDomains

	if req.OauthRefreshInterval != "" {
		shrCmd = append(shrCmd, "--oauth-refresh-interval", req.OauthRefreshInterval)
	}

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

	logrus.Infof("executing '%v'", shrCmd)

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
			logrus.Errorf("error joining: %v", err)
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
	for _, nssIn := range req.NamespaceSelections {
		out.NamespaceSelections = append(out.NamespaceSelections, NamespaceSelection{NamespaceToken: nssIn.NamespaceToken, Name: nssIn.Name})
	}
	if shareToken, frontendEndpoints, err := i.agent.SharePublic(out); err == nil {
		return &agentGrpc.SharePublicResponse{Token: shareToken, FrontendEndpoints: frontendEndpoints}, nil
	} else {
		return nil, err
	}
}
