package main

import (
	"github.com/michaelquigley/df/dd"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/controller/config"
	"github.com/openziti/zrok/tui"
	"github.com/spf13/cobra"
)

func init() {
	controllerCmd.cmd.AddCommand(newControllerValidateCommand().cmd)
}

type controllerValidateCommand struct {
	cmd *cobra.Command
}

func newControllerValidateCommand() *controllerValidateCommand {
	cmd := &cobra.Command{
		Use:   "validate <configPath>",
		Short: "Validate a zrok controller configuration document",
		Args:  cobra.ExactArgs(1),
	}
	command := &controllerValidateCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *controllerValidateCommand) run(_ *cobra.Command, args []string) {
	cfg, err := config.LoadConfig(args[0])
	if err != nil {
		tui.Error("controller config validation failed", err)
	}
	dl.Info(dd.MustInspect(cfg))
}
