package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/caddyserver/caddy/v2"
	_ "github.com/caddyserver/caddy/v2/modules/standard"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/openziti/zrok/environment"
	"net"
	"strings"
)

func init() {
	caddy.RegisterNetwork("zrok", newZrokListener)
}

type zrokListener struct {
	zctx  ziti.Context
	share string
	l     edge.Listener
}

func (l *zrokListener) String() string {
	return fmt.Sprintf("zrok/%s", l.share)
}

func (l *zrokListener) Network() string {
	return "zrok"
}

func (l *zrokListener) Accept() (net.Conn, error) {
	return l.l.Accept()
}

func (l *zrokListener) Close() error {
	_ = l.l.Close()
	l.zctx.Close()
	return nil
}

func (l *zrokListener) Addr() net.Addr {
	return l
}

func newZrokListener(ctx context.Context, _ string, addr string, cfg net.ListenConfig) (any, error) {
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
	l := &zrokListener{
		zctx:  zctx,
		share: shrToken,
		l:     conn,
	}
	return l, nil
}

var (
	_ net.Addr     = (*zrokListener)(nil)
	_ net.Listener = (*zrokListener)(nil)
)
