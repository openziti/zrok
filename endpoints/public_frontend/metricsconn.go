package public_frontend

import (
	"net"
	"time"
)

type metricsConn struct {
	id      string
	conn    net.Conn
	updates chan metricsUpdate
}

func newMetricsConn(id string, conn net.Conn, updates chan metricsUpdate) *metricsConn {
	return &metricsConn{id, conn, updates}
}

func (mc *metricsConn) Read(b []byte) (n int, err error) {
	n, err = mc.conn.Read(b)
	mc.updates <- metricsUpdate{mc.id, int64(n), 0}
	return n, err
}

func (mc *metricsConn) Write(b []byte) (n int, err error) {
	n, err = mc.conn.Write(b)
	mc.updates <- metricsUpdate{mc.id, 0, int64(n)}
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
