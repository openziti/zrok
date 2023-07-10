package sdk

import (
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/zrokdir"
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

func loadEnvironment(request *ShareRequest) (*zrokdir.ZrokDir, error) {
	zrd, err := zrokdir.Load()
	if err != nil {
		return nil, errors.Wrap(err, "error loading zrokdir")
	}
	return zrd, nil
}

func createShare(zrd *zrokdir.ZrokDir, req *rest_model_zrok.ShareRequest) (*Share, error) {
	return nil, nil
}
