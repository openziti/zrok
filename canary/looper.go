package canary

import (
	"bytes"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/openziti/zrok/environment/env_core"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"math/rand"
	"net/http"
	"time"
)

type PublicHttpLooper struct {
	id       int
	frontend string
	opt      *LooperOptions
	root     env_core.Root
	shr      *sdk.Share
	listener edge.Listener
	done     chan struct{}
}

func NewPublicHttpLooper(id int, frontend string, opt *LooperOptions, root env_core.Root) *PublicHttpLooper {
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
		logrus.Fatalf("#%d error starting: %v", l.id, err)
	}

	if err := l.bindListener(); err != nil {
		logrus.Fatalf("#%d error binding listener: %v", l.id, err)
	}

	l.dwell()

	logrus.Infof("completed #%d", l.id)
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

func (l *PublicHttpLooper) bindListener() error {
	zif, err := l.root.ZitiIdentityNamed(l.root.EnvironmentIdentityName())
	if err != nil {
		return errors.Wrapf(err, "#%d error getting identity", l.id)
	}
	zcfg, err := ziti.NewConfigFromFile(zif)
	if err != nil {
		return errors.Wrapf(err, "#%d error loading ziti config", l.id)
	}
	options := ziti.ListenOptions{
		ConnectTimeout:               5 * time.Minute,
		WaitForNEstablishedListeners: 1,
	}
	zctx, err := ziti.NewContext(zcfg)
	if err != nil {
		return errors.Wrapf(err, "#%d error creating ziti context", l.id)
	}

	if l.listener, err = zctx.ListenWithOptions(l.shr.Token, &options); err != nil {
		return errors.Wrapf(err, "#%d error binding listener", l.id)
	}

	go func() {
		if err := http.Serve(l.listener, l); err != nil {
			logrus.Errorf("#%d error starting http listener: %v", l.id, err)
		}
	}()

	return nil
}

func (l *PublicHttpLooper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	buf := new(bytes.Buffer)
	io.Copy(buf, r.Body)
	w.Write(buf.Bytes())
}

func (l *PublicHttpLooper) dwell() {
	dwell := l.opt.MinDwell.Milliseconds()
	dwelta := l.opt.MaxDwell.Milliseconds() - l.opt.MinDwell.Milliseconds()
	if dwelta > 0 {
		dwell = int64(rand.Intn(int(dwelta)) + int(l.opt.MinDwell.Milliseconds()))
	}
	time.Sleep(time.Duration(dwell) * time.Millisecond)
}

func (l *PublicHttpLooper) shutdown() error {
	if l.listener != nil {
		if err := l.listener.Close(); err != nil {
			logrus.Errorf("#%d error closing listener: %v", l.id, err)
		}
	}

	if err := sdk.DeleteShare(l.root, l.shr); err != nil {
		return errors.Wrapf(err, "#%d error deleting share", l.id)
	}

	return nil
}
