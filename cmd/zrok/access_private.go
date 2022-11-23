package main

import "github.com/spf13/cobra"

type accessPrivateCommand struct {
	cmd *cobra.Command
}

func newAccessPrivateCommand() *accessPrivateCommand {
	cmd := &cobra.Command{
		Use:   "private <serviceToken>",
		Short: "Create a private frontend to access a service",
		Args:  cobra.ExactArgs(1),
	}
	command := &accessPrivateCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *accessPrivateCommand) run(_ *cobra.Command, args []string) {
}
