package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/michaelquigley/df/dl"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/sirupsen/logrus"
)

func init() {
	pfxlog.GlobalInit(logrus.WarnLevel, pfxlog.DefaultOptions())
	dl.Init(dl.DefaultOptions())
}

func main() {
	data, err := loadData()
	if err != nil {
		panic(err)
	}

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

	fmt.Printf("access your pastebin using 'pastefrom %v'\n", shr.Token)

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
			go handle(conn, data)
		} else {
			panic(err)
		}
	}
}

func loadData() ([]byte, error) {
	stat, _ := os.Stdin.Stat()
	if stat.Mode()&os.ModeCharDevice == 0 {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, err
		}
		return data, nil
	} else {
		return nil, errors.New("'copyto' requires input from stdin; direct your paste buffer into stdin")
	}
}

func handle(conn net.Conn, data []byte) {
	_, err := conn.Write(data)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
}
