package frontend

import (
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

type metricsConn struct {
	id   string
	conn net.Conn
}

func newMetricsConn(id string, conn net.Conn) *metricsConn {
	return &metricsConn{id, conn}
}

func (mc *metricsConn) Read(b []byte) (n int, err error) {
	n, err = mc.conn.Read(b)
	logrus.Infof("[%v] => %d", mc.id, n)
	return n, err
}

func (mc *metricsConn) Write(b []byte) (n int, err error) {
	n, err = mc.conn.Write(b)
	logrus.Infof("[%v] <= %d", mc.id, n)
	return n, err
}

func (mc *metricsConn) Close() error {
	return mc.conn.Close()
}

func (mc *metricsConn) LocalAddr() net.Addr {
	return mc.conn.LocalAddr()
}

func (mc *metricsConn) RemoteAddr() net.Addr {
	return mc.conn.RemoteAddr()
}

func (mc *metricsConn) SetDeadline(t time.Time) error {
	return mc.conn.SetDeadline(t)
}

func (mc *metricsConn) SetReadDeadline(t time.Time) error {
	return mc.conn.SetReadDeadline(t)
}

func (mc *metricsConn) SetWriteDeadline(t time.Time) error {
	return mc.conn.SetWriteDeadline(t)
}
