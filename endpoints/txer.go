package endpoints

import (
	"io"
	"net"

	"github.com/michaelquigley/df/dl"
)

const bufSz = 10240

func TXer(from, to net.Conn) {
	dl.Debugf("started '%v' -> '%v'", from.RemoteAddr(), to.RemoteAddr())
	defer dl.Debugf("exited '%v' -> '%v'", from.RemoteAddr(), to.RemoteAddr())

	buf := make([]byte, bufSz)
	for {
		if rxsz, err := from.Read(buf); err == nil {
			if txsz, err := to.Write(buf[:rxsz]); err == nil {
				if txsz != rxsz {
					dl.Errorf("short write '%v' -> '%v' (%d != %d)", from.RemoteAddr(), to.RemoteAddr(), txsz, rxsz)
					_ = to.Close()
					_ = from.Close()
					return
				}
			} else {
				dl.Errorf("write error '%v' -> '%v': %v", from.RemoteAddr(), to.RemoteAddr(), err)
				_ = to.Close()
				_ = from.Close()
				return
			}
		} else {
			if err != io.EOF {
				dl.Errorf("read error '%v' -> '%v': %v", from.RemoteAddr(), to.RemoteAddr(), err)
			}
			_ = to.Close()
			_ = from.Close()
			return
		}
	}
}
