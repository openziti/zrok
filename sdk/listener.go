package sdk

import (
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/openziti/zrok/environment/env_core"
	"github.com/pkg/errors"
	"time"
)

func NewListener(shrToken string, root env_core.Root) (edge.Listener, error) {
	return NewListenerWithOptions(shrToken, root, &ziti.ListenOptions{ConnectTimeout: 30 * time.Second, MaxConnections: 64})
}

func NewListenerWithOptions(shrToken string, root env_core.Root, opts *ziti.ListenOptions) (edge.Listener, error) {
	zif, err := root.ZitiIdentityNamed(root.EnvironmentIdentityName())
	if err != nil {
		return nil, errors.Wrap(err, "error getting ziti identity path")
	}

	zcfg, err := ziti.NewConfigFromFile(zif)
	if err != nil {
		return nil, errors.Wrap(err, "error loading ziti identity")
	}

	zctx, err := ziti.NewContext(zcfg)
	if err != nil {
		return nil, errors.Wrap(err, "error getting ziti context")
	}

	listener, err := zctx.ListenWithOptions(shrToken, opts)
	if err != nil {
		return nil, errors.Wrap(err, "error creating listener")
	}

	return listener, nil
}
