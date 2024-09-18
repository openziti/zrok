package main

import (
	"context"
	"fmt"
	"github.com/openziti/zrok/agent/agentClient"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/tui"
	"github.com/spf13/cobra"
)

func init() {
	agentAccessCmd.AddCommand(newAgentAccessPrivateCommand().cmd)
}

type agentAccessPrivateCommand struct {
	bindAddress     string
	responseHeaders []string
	cmd             *cobra.Command
}

func newAgentAccessPrivateCommand() *agentAccessPrivateCommand {
	cmd := &cobra.Command{
		Use:   "private <token>",
		Short: "Bind a private access in the zrok Agent",
		Args:  cobra.ExactArgs(1),
	}
	command := &agentAccessPrivateCommand{cmd: cmd}
	cmd.Flags().StringVarP(&command.bindAddress, "bind", "b", "127.0.0.1:9191", "The address to bind the private frontend")
	cmd.Flags().StringArrayVar(&command.responseHeaders, "response-header", []string{}, "Add a response header ('key:value')")
	cmd.Run = command.run
	return command
}

func (cmd *agentAccessPrivateCommand) run(_ *cobra.Command, args []string) {
	root, err := environment.LoadRoot()
	if err != nil {
		if !panicInstead {
			tui.Error("unable to load environment", err)
		}
		panic(err)
	}

	if !root.IsEnabled() {
		tui.Error("unable to load environment; did you 'zrok enable'?", nil)
	}

	client, conn, err := agentClient.NewClient(root)
	if err != nil {
		tui.Error("error connecting to agent", err)
	}
	defer conn.Close()

	acc, err := client.AccessPrivate(context.Background(), &agentGrpc.AccessPrivateRequest{
		Token:           args[0],
		BindAddress:     cmd.bindAddress,
		ResponseHeaders: cmd.responseHeaders,
	})
	if err != nil {
		tui.Error("error creating access", err)
	}

	fmt.Println(acc)
}
