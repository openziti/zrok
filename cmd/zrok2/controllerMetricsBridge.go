package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/michaelquigley/df/dd"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/config"
	"github.com/openziti/zrok/v2/controller/metrics"
	"github.com/spf13/cobra"
)

func init() {
	metricsCmd.AddCommand(newBridgeCommand().cmd)
}

type bridgeCommand struct {
	cmd *cobra.Command
}

func newBridgeCommand() *bridgeCommand {
	cmd := &cobra.Command{
		Use:   "bridge <configPath>",
		Short: "Start a zrok metrics bridge",
		Args:  cobra.ExactArgs(1),
	}
	command := &bridgeCommand{cmd}
	cmd.Run = command.run
	return command
}

func (cmd *bridgeCommand) run(_ *cobra.Command, args []string) {
	cfg, err := config.LoadConfig(args[0])
	if err != nil {
		panic(err)
	}
	dl.Info(dd.MustInspect(cfg))

	bridge, err := metrics.NewBridge(cfg.Bridge)
	if err != nil {
		panic(err)
	}
	if _, err = bridge.Start(); err != nil {
		panic(err)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		bridge.Stop()
		os.Exit(0)
	}()

	for {
		time.Sleep(24 * 60 * time.Minute)
	}
}
