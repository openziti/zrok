package sdk

import (
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
