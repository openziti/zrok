package util

import (
	"fmt"
	"net"
)

func AutoListener(protocol, address string, startPort, endPort uint16) (net.Listener, error) {
	for i := startPort; i <= endPort; i++ {
		l, err := net.Listen(protocol, fmt.Sprintf("%s:%d", address, i))
		if err != nil {
			continue
		}
		return l, nil
	}
	return nil, fmt.Errorf("no listener found in range")
}

func AutoListenerAddress(protocol, address string, startPort, endPort uint16) (string, error) {
	listener, err := AutoListener(protocol, address, startPort, endPort)
	if err != nil {
		return "", err
	}
	autoAddress := listener.Addr().String()
	if err := listener.Close(); err != nil {
		return "", err
	}
	return autoAddress, nil
}
