package agent

import (
	"time"

	"github.com/openziti/zrok/v2/sdk/golang/sdk"
)

type SharePrivateRequest struct {
	Target            string   `dd:"target"`
	PrivateShareToken string   `dd:"private_share_token"`
	BackendMode       string   `dd:"backend_mode"`
	Insecure          bool     `dd:"insecure"`
	Closed            bool     `dd:"closed"`
	AccessGrants      []string `dd:"access_grants"`
}

func (spr *SharePrivateRequest) hasReservedToken() bool {
	return spr.PrivateShareToken != ""
}

type NameSelection struct {
	NamespaceToken string `dd:"namespace_token"`
	Name           string `dd:"name"`
}

type SharePublicRequest struct {
	Target               string          `dd:"target"`
	BasicAuth            []string        `dd:"basic_auth"`
	NameSelections       []NameSelection `dd:"name_selections"`
	BackendMode          string          `dd:"backend_mode"`
	Insecure             bool            `dd:"insecure"`
	OauthProvider        string          `dd:"oauth_provider"`
	OauthEmailDomains    []string        `dd:"oauth_email_domains"`
	OauthRefreshInterval string          `dd:"oauth_refresh_interval"`
	Closed               bool            `dd:"closed"`
	AccessGrants         []string        `dd:"access_grants"`
	// BasicAuthFile is a reference: the agent reads the users from it at each create, and the saved request
	// never holds them. empty means absent; the CLI refuses an empty path before it reaches the agent.
	BasicAuthFile string `dd:"basic_auth_file"`
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
	nameSelections            []NameSelection
	shareMode                 sdk.ShareMode
	backendMode               sdk.BackendMode
	insecure                  bool
	oauthProvider             string
	oauthEmailAddressPatterns []string
	oauthCheckInterval        time.Duration
	closed                    bool
	accessGrants              []string
}
