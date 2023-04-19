package main

import (
	"golang.org/x/sys/windows"
)

func newSignalHandler() chan os.Signal {
	signalHandler := make(chan os.Signal, 1)
	signal.Notify(signalHandler, windows.SIGINT, windows.SIGTERM, windows.SIGHUP, windows.SIGQUIT)
	return signalHandler
}
