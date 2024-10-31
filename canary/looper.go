package canary

import (
	"github.com/openziti/zrok/environment/env_core"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/sirupsen/logrus"
)

type PublicHttpLooper struct {
	id       int
	frontend string
	root     env_core.Root
	shr      *sdk.Share
	done     chan struct{}
}

func NewPublicHttpLooper(id int, frontend string, root env_core.Root) *PublicHttpLooper {
	return &PublicHttpLooper{
		id:       id,
		frontend: frontend,
		root:     root,
		done:     make(chan struct{}),
	}
}

func (l *PublicHttpLooper) Run() {
	defer close(l.done)
	defer logrus.Infof("stopping #%d", l.id)
	logrus.Infof("starting #%d", l.id)

	if err := l.startup(); err != nil {
		logrus.Fatalf("error starting #%d: %v", l.id, err)
	}

	logrus.Infof("#%d complete")
	if err := l.shutdown(); err != nil {
		logrus.Fatalf("error shutting down #%d: %v", l.id, err)
	}
}

func (l *PublicHttpLooper) startup() error {
	shr, err := sdk.CreateShare(l.root, &sdk.ShareRequest{
		ShareMode:      sdk.PublicShareMode,
		BackendMode:    sdk.ProxyBackendMode,
		Target:         "canary.PublicHttpLooper",
		Frontends:      []string{l.frontend},
		PermissionMode: sdk.ClosedPermissionMode,
	})
	if err != nil {
		return err
	}
	l.shr = shr
	logrus.Infof("#%d allocated share '%v'", l.id, l.shr)

	return nil
}

func (l *PublicHttpLooper) shutdown() error {
	return nil
}
