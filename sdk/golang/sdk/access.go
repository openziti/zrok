package sdk

import (
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti/zrok/environment/env_core"
	"github.com/openziti/zrok/rest_client_zrok/share"
	"github.com/pkg/errors"
)

func CreateAccess(root env_core.Root, request *AccessRequest) (*Access, error) {
	if !root.IsEnabled() {
		return nil, errors.New("environment is not enabled; enable with 'zrok enable' first!")
	}

	out := share.NewAccessParams()
	out.Body.ShareToken = request.ShareToken
	out.Body.EnvZID = root.Environment().ZitiIdentity

	zrok, err := root.Client()
	if err != nil {
		return nil, errors.Wrap(err, "error getting zrok client")
	}
	auth := httptransport.APIKeyAuth("X-TOKEN", "header", root.Environment().AccountToken)

	in, err := zrok.Share.Access(out, auth)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create access")
	}

	return &Access{Token: in.Payload.FrontendToken, ShareToken: request.ShareToken, BackendMode: BackendMode(in.Payload.BackendMode)}, nil
}

func DeleteAccess(root env_core.Root, acc *Access) error {
	out := share.NewUnaccessParams()
	out.Body.FrontendToken = acc.Token
	out.Body.ShareToken = acc.ShareToken
	out.Body.EnvZID = root.Environment().ZitiIdentity

	zrok, err := root.Client()
	if err != nil {
		return errors.Wrap(err, "error getting zrok client")
	}
	auth := httptransport.APIKeyAuth("X-TOKEN", "header", root.Environment().AccountToken)

	_, err = zrok.Share.Unaccess(out, auth)
	if err != nil {
		return errors.Wrap(err, "error deleting access")
	}

	return nil
}
