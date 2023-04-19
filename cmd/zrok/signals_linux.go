package main

import (
	"os"
	"os/signal"

	"golang.org/x/sys/unix"
)

func newSignalHandler() chan os.Signal {
	signalHandler := make(chan os.Signal, 1)
	signal.Notify(signalHandler, unix.SIGINT, unix.SIGTERM, unix.SIGHUP, unix.SIGQUIT)
	return signalHandler
}
