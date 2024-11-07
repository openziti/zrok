package util

import (
	"fmt"
	"net"
)

func AutoListener(address string, startPort, endPort uint16) (net.Listener, error) {
	for i := startPort; i <= endPort; i++ {
		l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", address, i))
		if err != nil {
			continue
		}
		return l, nil
	}
	return nil, fmt.Errorf("no listener found in range")
}
