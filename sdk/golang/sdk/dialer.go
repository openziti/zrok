package sdk

import (
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/openziti/zrok/v2/environment/env_core"
	"github.com/pkg/errors"
	"time"
)

func NewDialer(shrToken string, root env_core.Root) (edge.Conn, error) {
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

	conn, err := zctx.DialWithOptions(shrToken, &ziti.DialOptions{ConnectTimeout: 30 * time.Second})
	if err != nil {
		return nil, errors.Wrapf(err, "error dialing '%v'", shrToken)
	}

	return conn, nil
}
