package tcpTunnel

import (
	"github.com/sirupsen/logrus"
	"io"
	"net"
)

const bufSz = 10240

func txer(from, to net.Conn) {
	logrus.Debugf("started '%v' -> '%v'", from.RemoteAddr(), to.RemoteAddr())
	defer logrus.Debugf("exited '%v' -> '%v'", from.RemoteAddr(), to.RemoteAddr())

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
			if err != io.EOF {
				logrus.Errorf("read error '%v' -> '%v': %v", from.RemoteAddr(), to.RemoteAddr(), err)
			}
			_ = to.Close()
			_ = from.Close()
			return
		}
	}
}
