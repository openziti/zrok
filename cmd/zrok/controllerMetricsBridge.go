package main

import (
	"github.com/michaelquigley/cf"
	"github.com/openziti/zrok/controller/config"
	"github.com/openziti/zrok/controller/env"
	"github.com/openziti/zrok/controller/metrics2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
	"time"
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
	logrus.Infof(cf.Dump(cfg, env.GetCfOptions()))

	bridge, err := metrics2.NewBridge(cfg.Bridge)
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
