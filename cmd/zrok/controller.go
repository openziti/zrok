package main

import (
	"github.com/michaelquigley/df"
	"github.com/openziti/zrok/controller"
	"github.com/openziti/zrok/controller/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var controllerCmd *controllerCommand

var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Metrics related commands",
}

func init() {
	controllerCmd = newControllerCommand()
	controllerCmd.cmd.AddCommand(metricsCmd)
	rootCmd.AddCommand(controllerCmd.cmd)
}

type controllerCommand struct {
	cmd *cobra.Command
}

func newControllerCommand() *controllerCommand {
	cmd := &cobra.Command{
		Use:     "controller <configPath>",
		Short:   "Start a zrok controller",
		Aliases: []string{"ctrl"},
		Args:    cobra.ExactArgs(1),
	}
	command := &controllerCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *controllerCommand) run(_ *cobra.Command, args []string) {
	cfg, err := config.LoadConfig(args[0])
	if err != nil {
		panic(err)
	}
	logrus.Info(df.Inspect(cfg))
	if err := controller.Run(cfg); err != nil {
		panic(err)
	}
}
