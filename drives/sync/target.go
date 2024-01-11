package sync

import (
	"github.com/openziti/zrok/environment/env_core"
	"github.com/pkg/errors"
	"net/url"
)

func TargetForURL(url *url.URL, root env_core.Root) (Target, error) {
	switch url.Scheme {
	case "file":
		return NewFilesystemTarget(&FilesystemTargetConfig{Root: url.Path}), nil

	case "zrok":
		return NewZrokTarget(&ZrokTargetConfig{URL: url, Root: root})

	case "http", "https":
		return NewWebDAVTarget(&WebDAVTargetConfig{URL: url, Username: "", Password: ""})

	default:
		return nil, errors.Errorf("unknown URL scheme '%v'", url.Scheme)
	}
}
