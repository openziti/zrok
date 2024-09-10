package agent

import (
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"time"
)

type share struct {
	token  string
	target string

	basicAuth                 []string
	frontendSelection         []string
	shareMode                 sdk.ShareMode
	backendMode               sdk.BackendMode
	reserved                  bool
	insecure                  bool
	oauthProvider             string
	oauthEmailAddressPatterns []string
	oauthCheckInterval        time.Duration
	closed                    bool
	accessGrants              []string

	handler backendHandler
}

type access struct {
	token string

	bindAddress     string
	responseHeaders []string
}

type agentGrpcImpl struct {
	agentGrpc.UnimplementedAgentServer
	a *Agent
}

type backendHandler interface {
	Run() error
	Stop() error
}
