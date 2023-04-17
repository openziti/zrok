package main

import (
	"github.com/openziti/zrok/endpoints/tunnelFrontend"
	"github.com/spf13/cobra"
	"time"
)

func init() {
	accessPrivateCmd.cmd.AddCommand(newAccessPrivateTunnelCommand().cmd)
}

type accessPrivateTunnelCommand struct {
	bindAddress string
	cmd         *cobra.Command
}

func newAccessPrivateTunnelCommand() *accessPrivateTunnelCommand {
	cmd := &cobra.Command{
		Use:   "tunnel <shareToken>",
		Short: "Create a private tunnel frontend to access a share",
		Args:  cobra.ExactArgs(1),
	}
	command := &accessPrivateTunnelCommand{cmd: cmd}
	cmd.Flags().StringVarP(&command.bindAddress, "bind", "b", "tcp:127.0.0.1:9191", "The address to bind the private tunnel")
	cmd.Run = command.run
	return command
}

func (cmd *accessPrivateTunnelCommand) run(_ *cobra.Command, args []string) {
	fe, err := tunnelFrontend.NewFrontend(&tunnelFrontend.Config{
		BindAddress:  cmd.bindAddress,
		IdentityName: "backend",
		ShrToken:     args[0],
	})
	if err != nil {
		panic(err)
	}
	if err := fe.Run(); err != nil {
		panic(err)
	}
	for {
		time.Sleep(50)
	}
}
