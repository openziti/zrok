package proxy

import (
	"context"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/sirupsen/logrus"
	"net"
	"strings"
)

type ZitiDialContext struct {
	Context ziti.Context
}

func (self *ZitiDialContext) Dial(_ context.Context, _ string, addr string) (net.Conn, error) {
	service := strings.Split(addr, ":")[0] // ignore :port (we get passed 'host:port')
	_, found := self.Context.GetService(service)
	if !found {
		logrus.Infof("service '%v' not cached; refreshing", service)
		self.Context.RefreshServices()
	}
	return self.Context.Dial(service)
}
