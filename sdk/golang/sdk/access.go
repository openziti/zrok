package sdk

import (
	"context"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/environment/env_core"
	"github.com/openziti/zrok/v2/rest_client_zrok/share"
	"github.com/pkg/errors"
)

func CreateAccess(root env_core.Root, request *AccessRequest) (*Access, error) {
	if !root.IsEnabled() {
		return nil, errors.New("environment is not enabled; enable with 'zrok2 enable' first!")
	}

	out := share.NewAccessParams()
	out.Body.ShareToken = request.ShareToken
	out.Body.EnvZID = root.Environment().ZitiIdentity
	if request.BindAddress != "" {
		out.Body.BindAddress = request.BindAddress
		dl.Infof("requesting bind address '%v'", out.Body.BindAddress)
	}

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
	return deleteAccess(nil, root, acc)
}

// DeleteAccessContext bounds initialization and deletion with the same context.
func DeleteAccessContext(ctx context.Context, root env_core.Root, acc *Access) error {
	if ctx == nil {
		return errors.New("missing access deletion context")
	}
	return deleteAccess(ctx, root, acc)
}

func deleteAccess(ctx context.Context, root env_core.Root, acc *Access) error {
	out := share.NewUnaccessParams()
	out.Context = ctx
	out.Body.FrontendToken = acc.Token
	out.Body.ShareToken = acc.ShareToken
	out.Body.EnvZID = root.Environment().ZitiIdentity

	zrok, err := shareClient(ctx, root)
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
