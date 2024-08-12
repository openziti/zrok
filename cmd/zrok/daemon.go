package main

import (
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/tui"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newDaemonCommand().cmd)
}

type daemonCommand struct {
	cmd *cobra.Command
}

func newDaemonCommand() *daemonCommand {
	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Launch a zrok daemon",
		Args:  cobra.NoArgs,
	}
	command := &daemonCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *daemonCommand) run(_ *cobra.Command, _ []string) {
	root, err := environment.LoadRoot()
	if err != nil {
		tui.Error("error loading zrokdir", err)
	}

	if !root.IsEnabled() {
		tui.Error("unable to load environment; did you 'zrok enable'?", nil)
	}
}
