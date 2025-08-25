package endpoints

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/caddyserver/caddy/v2"
	_ "github.com/caddyserver/caddy/v2/modules/standard"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/openziti/zrok/environment"
)

func init() {
	caddy.RegisterNetwork("zrok", NewZrokListener)
}

type ZrokListener struct {
	zctx  ziti.Context
	share string
	l     edge.Listener
}

func (l *ZrokListener) String() string {
	return fmt.Sprintf("zrok/%s", l.share)
}

func (l *ZrokListener) Network() string {
	return "zrok"
}

func (l *ZrokListener) Accept() (net.Conn, error) {
	return l.l.Accept()
}

func (l *ZrokListener) Close() error {
	_ = l.l.Close()
	l.zctx.Close()
	return nil
}

func (l *ZrokListener) Addr() net.Addr {
	return l
}

func NewZrokListener(_ context.Context, _ string, addr string, _ string, _ uint, _ net.ListenConfig) (any, error) {
	shrToken := strings.Split(addr, ":")[0]
	env, err := environment.LoadRoot()
	if err != nil {
		return nil, err
	}
	if !env.IsEnabled() {
		return nil, errors.New("environment not enabled")
	}
	zif, err := env.ZitiIdentityNamed(env.EnvironmentIdentityName())
	if err != nil {
		return nil, err
	}
	zctx, err := ziti.NewContextFromFile(zif)
	if err != nil {
		return nil, err
	}
	conn, err := zctx.Listen(shrToken)
	if err != nil {
		return nil, err
	}
	l := &ZrokListener{
		zctx:  zctx,
		share: shrToken,
		l:     conn,
	}
	return l, nil
}

var (
	_ net.Addr     = (*ZrokListener)(nil)
	_ net.Listener = (*ZrokListener)(nil)
)
