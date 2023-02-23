package main

import (
	"github.com/michaelquigley/cf"
	"github.com/openziti/zrok/controller"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newUsageCommand().cmd)
}

type usageCommand struct {
	cmd *cobra.Command
}

func newUsageCommand() *usageCommand {
	cmd := &cobra.Command{
		Use:   "usage <configPath>",
		Short: "Start a zrok metrics agent",
		Args:  cobra.ExactArgs(1),
	}
	command := &usageCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *usageCommand) run(_ *cobra.Command, args []string) {
	cfg, err := controller.LoadConfig(args[0])
	if err != nil {
		panic(err)
	}
	logrus.Infof(cf.Dump(cfg, cf.DefaultOptions()))

	if err := controller.RunUsageAgent(cfg); err != nil {
		panic(err)
	}
}
