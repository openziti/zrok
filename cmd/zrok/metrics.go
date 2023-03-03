package main

import (
	"github.com/michaelquigley/cf"
	"github.com/openziti/zrok/controller/metrics"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
	cfg, err := metrics.LoadConfig(args[0])
	if err != nil {
		panic(err)
	}
	logrus.Infof(cf.Dump(cfg, metrics.GetCfOptions()))

	if err := metrics.Run(cfg); err != nil {
		panic(err)
	}
}
