package daemon

import (
	"github.com/openziti/zrok/sdk/golang/sdk"
	"time"
)

type share struct {
	token string

	basicAuth                 []string
	frontendSelection         []string
	backendMode               sdk.BackendMode
	insecure                  bool
	oauthProvider             string
	oauthEmailAddressPatterns []string
	oauthCheckInterval        time.Duration
	closed                    bool
	accessGrants              []string
}

type access struct {
	token string

	bindAddress     string
	responseHeaders []string
}
