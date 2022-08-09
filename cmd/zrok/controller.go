package main

import (
	"github.com/openziti-test-kitchen/zrok/controller"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newControllerCommand().cmd)
}

type controllerCommand struct {
	dbPath string
	cmd    *cobra.Command
}

func newControllerCommand() *controllerCommand {
	cmd := &cobra.Command{
		Use:     "controller <configPath>",
		Short:   "Start a zrok controller",
		Aliases: []string{"ctrl"},
		Args:    cobra.ExactArgs(1),
	}
	ccmd := &controllerCommand{
		cmd: cmd,
	}
	cmd.Run = ccmd.run
	return ccmd
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
