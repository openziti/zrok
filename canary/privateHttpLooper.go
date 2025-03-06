package canary

import (
	"bytes"
	"context"
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
	"net"
	"net/http"
	"time"
)

type PrivateHttpLooper struct {
	id       uint
	acc      *sdk.Access
	opt      *LooperOptions
	root     env_core.Root
	shr      *sdk.Share
	listener edge.Listener
	abort    bool
	done     chan struct{}
	results  *LooperResults
}

func NewPrivateHttpLooper(id uint, opt *LooperOptions, root env_core.Root) *PrivateHttpLooper {
	return &PrivateHttpLooper{
		id:      id,
		opt:     opt,
		root:    root,
		done:    make(chan struct{}),
		results: &LooperResults{},
	}
}

func (l *PrivateHttpLooper) Run() {
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

func (l *PrivateHttpLooper) Abort() {
	l.abort = true
}

func (l *PrivateHttpLooper) Done() <-chan struct{} {
	return l.done
}

func (l *PrivateHttpLooper) Results() *LooperResults {
	return l.results
}

func (l *PrivateHttpLooper) startup() error {
	shr, err := sdk.CreateShare(l.root, &sdk.ShareRequest{
		ShareMode:      sdk.PrivateShareMode,
		BackendMode:    sdk.ProxyBackendMode,
		Target:         "canary.PrivateHttpLooper",
		PermissionMode: sdk.ClosedPermissionMode,
	})
	if err != nil {
		return err
	}
	l.shr = shr

	acc, err := sdk.CreateAccess(l.root, &sdk.AccessRequest{
		ShareToken: shr.Token,
	})
	if err != nil {
		return err
	}
	l.acc = acc

	logrus.Infof("#%d allocated share '%v', allocated frontend '%v'", l.id, shr.Token, acc.Token)

	return nil
}

func (l *PrivateHttpLooper) bind() error {
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

func (l *PrivateHttpLooper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	buf := new(bytes.Buffer)
	io.Copy(buf, r.Body)
	w.Write(buf.Bytes())
}

func (l *PrivateHttpLooper) dwell() {
	dwell := l.opt.MinDwell.Milliseconds()
	dwelta := l.opt.MaxDwell.Milliseconds() - l.opt.MinDwell.Milliseconds()
	if dwelta > 0 {
		dwell = int64(rand.Intn(int(dwelta)) + int(l.opt.MinDwell.Milliseconds()))
	}
	time.Sleep(time.Duration(dwell) * time.Millisecond)
}

type connDialer struct {
	c net.Conn
}

func (cd connDialer) Dial(_ context.Context, network, addr string) (net.Conn, error) {
	return cd.c, nil
}

func (l *PrivateHttpLooper) iterate() {
	l.results.StartTime = time.Now()
	defer func() { l.results.StopTime = time.Now() }()

	for i := uint(0); i < l.opt.Iterations; i++ {
		if i > 0 && i%l.opt.StatusInterval == 0 {
			logrus.Infof("#%d: iteration %d", l.id, i)
		}

		conn, err := sdk.NewDialer(l.shr.Token, l.root)
		if err != nil {
			logrus.Errorf("#%d: error dialing: %v", l.id, err)
			l.results.Errors++
			time.Sleep(1 * time.Second)
			continue
		}

		payloadSize := l.opt.MaxPayload
		payloadRange := l.opt.MaxPayload - l.opt.MinPayload
		if payloadRange > 0 {
			payloadSize = (rand.Uint64() % payloadRange) + l.opt.MinPayload
		}
		outPayload := make([]byte, payloadSize)
		cryptorand.Read(outPayload)
		outBase64 := base64.StdEncoding.EncodeToString(outPayload)

		if req, err := http.NewRequest("POST", "http://"+l.shr.Token, bytes.NewBufferString(outBase64)); err == nil {
			client := &http.Client{Timeout: l.opt.Timeout, Transport: &http.Transport{DialContext: connDialer{conn}.Dial}}
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

		if err := conn.Close(); err != nil {
			logrus.Errorf("#%d: error closing connection: %v", l.id, err)
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

func (l *PrivateHttpLooper) shutdown() {
	if l.listener != nil {
		if err := l.listener.Close(); err != nil {
			logrus.Errorf("#%d error closing listener: %v", l.id, err)
		}
	}

	if err := sdk.DeleteAccess(l.root, l.acc); err != nil {
		logrus.Errorf("#%d error deleting access '%v': %v", l.id, l.acc.Token, err)
	}

	if err := sdk.DeleteShare(l.root, l.shr); err != nil {
		logrus.Errorf("#%d error deleting share '%v': %v", l.id, l.shr.Token, err)
	}
}
