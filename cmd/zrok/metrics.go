package main

import (
	"github.com/michaelquigley/cf"
	"github.com/openziti/zrok/controller/config"
	"github.com/openziti/zrok/controller/env"
	"github.com/openziti/zrok/controller/metrics"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func init() {
	rootCmd.AddCommand(newMetricsCommand().cmd)
}

type metricsCommand struct {
	cmd *cobra.Command
}

func newMetricsCommand() *metricsCommand {
	cmd := &cobra.Command{
		Use:   "metrics <configPath>",
		Short: "Start a zrok metrics agent",
		Args:  cobra.ExactArgs(1),
	}
	command := &metricsCommand{cmd}
	cmd.Run = command.run
	return command
}

func (cmd *metricsCommand) run(_ *cobra.Command, args []string) {
	cfg, err := config.LoadConfig(args[0])
	if err != nil {
		panic(err)
	}
	logrus.Infof(cf.Dump(cfg, env.GetCfOptions()))

	ma, err := metrics.Run(cfg.Metrics, cfg.Store)
	if err != nil {
		panic(err)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		ma.Stop()
		ma.Join()
		os.Exit(0)
	}()

	for {
		time.Sleep(30 * time.Minute)
	}
}
