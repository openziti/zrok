package main

import (
	"github.com/openziti-test-kitchen/zrok/zrokdir"
	"github.com/spf13/cobra"
)

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
	env, err := zrokdir.LoadEnvironment()
	if err != nil {
		if !panicInstead {
			showError("unable to load environment; did you 'zrok enable'?", err)
		}
		panic(err)
	}
	zif, err := zrokdir.ZitiIdentityFile("backend")
	if err != nil {
		if !panicInstead {
			showError("unable to load ziti identity configuration", err)
		}
		panic(err)
	}
	if zif == "" {
		panic("never")
	}
	zrok, err := zrokdir.ZrokClient(env.ApiEndpoint)
	if err != nil {
		if !panicInstead {
			showError("unable to create zrok client", err)
		}
		panic(err)
	}
	if zrok == nil {
		panic("never")
	}
}
