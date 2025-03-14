package canary

import (
	"bytes"
	cryptorand "crypto/rand"
	"encoding/base64"
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
	id       uint
	frontend string
	opt      *LooperOptions
	root     env_core.Root
	shr      *sdk.Share
	listener edge.Listener
	abort    bool
	done     chan struct{}
	results  *LooperResults
}

func NewPublicHttpLooper(id uint, frontend string, opt *LooperOptions, root env_core.Root) *PublicHttpLooper {
	return &PublicHttpLooper{
		id:       id,
		frontend: frontend,
		opt:      opt,
		root:     root,
		done:     make(chan struct{}),
		results:  &LooperResults{},
	}
}

func (l *PublicHttpLooper) Run() {
	defer close(l.done)
	defer logrus.Infof("#%d stopping", l.id)
	defer l.shutdown()
	logrus.Infof("#%d starting", l.id)

	if err := l.startup(); err != nil {
		logrus.Fatalf("#%d error starting: %v", l.id, err)
	}

	if err := l.bind(); err != nil {
		logrus.Fatalf("#%d error binding: %v", l.id, err)
	}

	l.dwell()

	l.iterate()

	logrus.Infof("#%d completed", l.id)
}

func (l *PublicHttpLooper) Abort() {
	l.abort = true
}

func (l *PublicHttpLooper) Done() <-chan struct{} {
	return l.done
}

func (l *PublicHttpLooper) Results() *LooperResults {
	return l.results
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

	logrus.Infof("#%d allocated share '%v'", l.id, l.shr.Token)

	return nil
}

func (l *PublicHttpLooper) bind() error {
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
			logrus.Errorf("#%d error in http listener: %v", l.id, err)
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

func (l *PublicHttpLooper) iterate() {
	l.results.StartTime = time.Now()
	defer func() { l.results.StopTime = time.Now() }()

	for i := uint(0); i < l.opt.Iterations && !l.abort; i++ {
		if i > 0 && i%l.opt.StatusInterval == 0 {
			logrus.Infof("#%d: iteration %d", l.id, i)
		}

		payloadSize := l.opt.MaxPayload
		payloadRange := l.opt.MaxPayload - l.opt.MinPayload
		if payloadRange > 0 {
			payloadSize = (rand.Uint64() % payloadRange) + l.opt.MinPayload
		}
		outPayload := make([]byte, payloadSize)
		cryptorand.Read(outPayload)
		outBase64 := base64.StdEncoding.EncodeToString(outPayload)

		if req, err := http.NewRequest("POST", l.shr.FrontendEndpoints[0], bytes.NewBufferString(outBase64)); err == nil {
			client := &http.Client{Timeout: l.opt.Timeout}
			if resp, err := client.Do(req); err == nil {
				if resp.StatusCode != 200 {
					logrus.Errorf("#%d: unexpected status code: %v", l.id, resp.StatusCode)
					l.results.Errors++
				}
				inPayload := new(bytes.Buffer)
				io.Copy(inPayload, resp.Body)
				inBase64 := inPayload.String()
				if inBase64 != outBase64 {
					logrus.Errorf("#%d: payload mismatch", l.id)
					l.results.Mismatches++
				} else {
					l.results.Bytes += uint64(len(outBase64))
					logrus.Debugf("#%d: payload match", l.id)
				}
			} else {
				logrus.Errorf("#%d: error: %v", l.id, err)
				l.results.Errors++
			}
		} else {
			logrus.Errorf("#%d: error creating request: %v", l.id, err)
			l.results.Errors++
		}

		pacingMs := l.opt.MaxPacing.Milliseconds()
		pacingDelta := l.opt.MaxPacing.Milliseconds() - l.opt.MinPacing.Milliseconds()
		if pacingDelta > 0 {
			pacingMs = (rand.Int63() % pacingDelta) + l.opt.MinPacing.Milliseconds()
			time.Sleep(time.Duration(pacingMs) * time.Millisecond)
		}

		l.results.Loops++
	}
}

func (l *PublicHttpLooper) shutdown() {
	if l.listener != nil {
		if err := l.listener.Close(); err != nil {
			logrus.Errorf("#%d error closing listener: %v", l.id, err)
		}
	}

	if err := sdk.DeleteShare(l.root, l.shr); err != nil {
		logrus.Errorf("#%d error deleting share '%v': %v", l.id, l.shr.Token, err)
	}
}
