package main

import (
	"github.com/openziti-test-kitchen/zrok/controller"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newControllerCommand().cmd)
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
	cfg, err := controller.LoadConfig(args[0])
	if err != nil {
		panic(err)
	}
	if err := controller.Run(cfg); err != nil {
		panic(err)
	}
}
