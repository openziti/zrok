package sync

import (
	"github.com/openziti/zrok/environment/env_core"
	"github.com/pkg/errors"
	"net/url"
	"strings"
)

func TargetForURL(url *url.URL, root env_core.Root, basicAuth string) (Target, error) {
	switch url.Scheme {
	case "file":
		return NewFilesystemTarget(&FilesystemTargetConfig{Root: url.Path}), nil

	case "zrok":
		return NewZrokTarget(&ZrokTargetConfig{URL: url, Root: root})

	case "http", "https":
		var username string
		var password string
		if basicAuth != "" {
			authTokens := strings.Split(basicAuth, ":")
			if len(authTokens) != 2 {
				return nil, errors.Errorf("invalid basic authentication (expect 'username:password')")
			}
			username = authTokens[0]
			password = authTokens[1]
		}
		return NewWebDAVTarget(&WebDAVTargetConfig{URL: url, Username: username, Password: password})

	default:
		return nil, errors.Errorf("unknown URL scheme '%v'", url.Scheme)
	}
}
