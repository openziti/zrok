package agent

import (
	"time"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/agent/proctree"
	"github.com/openziti/zrok/v2/cmd/zrok/subordinate"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
)

type SharePrivateRequest struct {
	Target            string   `json:"target"`
	PrivateShareToken string   `json:"private_share_token"`
	BackendMode       string   `json:"backend_mode"`
	Insecure          bool     `json:"insecure"`
	Closed            bool     `json:"closed"`
	AccessGrants      []string `json:"access_grants"`
}

func (spr *SharePrivateRequest) hasReservedToken() bool {
	return spr.PrivateShareToken != ""
}

type NameSelection struct {
	NamespaceToken string `json:"namespace_token"`
	Name           string `json:"name"`
}

type SharePublicRequest struct {
	Target               string          `json:"target"`
	BasicAuth            []string        `json:"basic_auth"`
	NameSelections       []NameSelection `json:"name_selections"`
	BackendMode          string          `json:"backend_mode"`
	Insecure             bool            `json:"insecure"`
	OauthProvider        string          `json:"oauth_provider"`
	OauthEmailDomains    []string        `json:"oauth_email_domains"`
	OauthRefreshInterval string          `json:"oauth_refresh_interval"`
	Closed               bool            `json:"closed"`
	AccessGrants         []string        `json:"access_grants"`
}

func (spr *SharePublicRequest) hasReservedName() bool {
	for _, ns := range spr.NameSelections {
		if ns.Name != "" {
			return true
		}
	}
	return false
}

type share struct {
	token                     string
	frontendEndpoints         []string
	target                    string
	basicAuth                 []string
	nameSelections            []NameSelection
	shareMode                 sdk.ShareMode
	backendMode               sdk.BackendMode
	insecure                  bool
	oauthProvider             string
	oauthEmailAddressPatterns []string
	oauthCheckInterval        time.Duration
	closed                    bool
	accessGrants              []string

	request          interface{}
	releaseRequested bool
	processExited    bool
	lastError        error

	process *proctree.Child
	sub     *subordinate.MessageHandler

	agent *Agent
}

func (s *share) monitor() {
	if err := proctree.WaitChild(s.process); err != nil {
		dl.ChannelLog(s.token).Error(err)
		s.lastError = err
	}
	s.processExited = true
	s.agent.rmShare <- s
}

