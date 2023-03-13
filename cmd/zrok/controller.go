package main

import (
	"github.com/michaelquigley/cf"
	"github.com/openziti/zrok/controller"
	"github.com/openziti/zrok/controller/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var controllerCmd *controllerCommand

func init() {
	controllerCmd = newControllerCommand()
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
	logrus.Infof(cf.Dump(cfg, cf.DefaultOptions()))
	if err := controller.Run(cfg); err != nil {
		panic(err)
	}
}
