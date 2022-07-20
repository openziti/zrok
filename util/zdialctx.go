package util

import (
	"context"
	"github.com/openziti/sdk-golang/ziti"
	"net"
	"strings"
)

type ZitiDialContext struct {
	Context ziti.Context
}

func (self *ZitiDialContext) Dial(_ context.Context, _ string, addr string) (net.Conn, error) {
	service := strings.Split(addr, ":")[0] // ignore :port (we get passed 'host:port')
	return self.Context.Dial(service)
}
