package sdk

import (
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/pkg/errors"
)

type Share struct {
	Token string
}

func NewShare(request *ShareRequest) (*Share, error) {
	switch request.ShareMode {
	case PrivateShareMode:
		return newPrivateShare(request)
	case PublicShareMode:
		return newPublicShare(request)
	default:
		return nil, errors.Errorf("unknown share mode '%v'", request.ShareMode)
	}
}

func newPrivateShare(request *ShareRequest) (*Share, error) {
	return nil, nil
}

func newPublicShare(request *ShareRequest) (*Share, error) {
	return nil, nil
}

func loadEnvironment(request *ShareRequest) (*environment.Root, error) {
	env, err := environment.Load()
	if err != nil {
		return nil, errors.Wrap(err, "error loading environment")
	}
	return env, nil
}

func createShare(zrd *environment.Root, req *rest_model_zrok.ShareRequest) (*Share, error) {
	return nil, nil
}
