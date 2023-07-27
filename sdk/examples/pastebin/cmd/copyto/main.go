package main

import (
	"fmt"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/sdk"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	root, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	shr, err := sdk.CreateShare(root, &sdk.ShareRequest{
		BackendMode: sdk.TcpTunnelBackendMode,
		ShareMode:   sdk.PrivateShareMode,
		Target:      "pastebin",
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("access your pastebin with: 'pastefrom %v'\n", shr.Token)

	listener, err := sdk.NewListener(shr.Token, root)
	if err != nil {
		panic(err)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		if err := sdk.DeleteShare(root, shr); err != nil {
			panic(err)
		}
		_ = listener.Close()
		os.Exit(0)
	}()

	for {
		if conn, err := listener.Accept(); err == nil {
			go handle(conn)
		} else {
			panic(err)
		}
	}
}

func handle(conn net.Conn) {
	_, err := conn.Write([]byte("hello from pastebin"))
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
}
