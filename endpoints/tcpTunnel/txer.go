package tcpTunnel

import (
	"github.com/sirupsen/logrus"
	"net"
)

const bufSz = 10240

func txer(from, to net.Conn) {
	logrus.Infof("started '%v' -> '%v'", from.RemoteAddr(), to.RemoteAddr())
	defer logrus.Warnf("exited '%v' -> '%v'", from.RemoteAddr(), to.RemoteAddr())

	buf := make([]byte, bufSz)
	for {
		if rxsz, err := from.Read(buf); err == nil {
			if txsz, err := to.Write(buf[:rxsz]); err == nil {
				if txsz != rxsz {
					logrus.Errorf("short write '%v' -> '%v' (%d != %d)", from.RemoteAddr(), to.RemoteAddr(), txsz, rxsz)
					_ = to.Close()
					_ = from.Close()
					return
				}
			} else {
				logrus.Errorf("write error '%v' -> '%v': %v", from.RemoteAddr(), to.RemoteAddr(), err)
				_ = to.Close()
				_ = from.Close()
				return
			}
		} else {
			logrus.Errorf("read error '%v' -> '%v': %v", from.RemoteAddr(), to.RemoteAddr(), err)
			_ = to.Close()
			_ = from.Close()
			return
		}
	}
}
