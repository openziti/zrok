package sdk

import (
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti/zrok/environment/env_core"
	restEnvironment "github.com/openziti/zrok/rest_client_zrok/environment"
	"github.com/pkg/errors"
)

func EnableEnvironment(root env_core.Root, request *EnableRequest) (*Environment, error) {
	zrok, err := root.Client()
	if err != nil {
		return nil, errors.Wrap(err, "could not create zrok client")
	}
	auth := httptransport.APIKeyAuth("X-TOKEN", "header", root.Environment().AccountToken)

	req := restEnvironment.NewEnableParams()
	req.Body.Description = request.Description
	req.Body.Host = request.Host

	resp, err := zrok.Environment.Enable(req, auth)
	if err != nil {
		return nil, err
	}

	return &Environment{
		Host:         request.Host,
		Description:  request.Description,
		ZitiIdentity: resp.Payload.Identity,
		ZitiConfig:   resp.Payload.Cfg,
	}, nil
}

func DisableEnvironment(env *Environment, root env_core.Root) error {
	zrok, err := root.Client()
	if err != nil {
		return errors.Wrap(err, "could not create zrok client")
	}
	auth := httptransport.APIKeyAuth("X-TOKEN", "header", root.Environment().AccountToken)

	req := restEnvironment.NewDisableParams()
	req.Body.Identity = env.ZitiIdentity

	_, err = zrok.Environment.Disable(req, auth)
	if err != nil {
		return err
	}

	return nil
}
