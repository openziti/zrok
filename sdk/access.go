package sdk

import (
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti/zrok/environment/env_core"
	"github.com/openziti/zrok/rest_client_zrok/share"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/pkg/errors"
)

func CreateAccess(root env_core.Root, request *AccessRequest) (*Access, error) {
	out := share.NewAccessParams()
	out.Body = &rest_model_zrok.AccessRequest{
		ShrToken: request.ShareToken,
		EnvZID:   root.Environment().Token,
	}

	zrok, err := root.Client()
	if err != nil {
		return nil, errors.Wrap(err, "error getting zrok client")
	}
	auth := httptransport.APIKeyAuth("X-TOKEN", "header", root.Environment().Token)

	in, err := zrok.Share.Access(out, auth)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create access")
	}

	return &Access{Token: in.Payload.FrontendToken, ShareToken: request.ShareToken, BackendMode: BackendMode(in.Payload.BackendMode)}, nil
}

func DeleteAccess(root env_core.Root, acc *Access) error {
	out := share.NewUnaccessParams()
	out.Body = &rest_model_zrok.UnaccessRequest{
		FrontendToken: acc.Token,
		ShrToken:      acc.ShareToken,
		EnvZID:        root.Environment().ZitiIdentity,
	}

	zrok, err := root.Client()
	if err != nil {
		return errors.Wrap(err, "error getting zrok client")
	}
	auth := httptransport.APIKeyAuth("X-TOKEN", "header", root.Environment().Token)

	_, err = zrok.Share.Unaccess(out, auth)
	if err != nil {
		return errors.Wrap(err, "error deleting access")
	}

	return nil
}
