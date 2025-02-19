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
	agentReleaseCmd.AddCommand(newAgentReleaseAccessCommand().cmd)
}

type agentReleaseAccessCommand struct {
	cmd *cobra.Command
}

func newAgentReleaseAccessCommand() *agentReleaseAccessCommand {
	cmd := &cobra.Command{
		Use:   "access <frontendToken>",
		Short: "Unbind an access from the zrok Agent",
		Args:  cobra.ExactArgs(1),
	}
	command := &agentReleaseAccessCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *agentReleaseAccessCommand) run(_ *cobra.Command, args []string) {
	root, err := environment.LoadRoot()
	if err != nil {
		if !panicInstead {
			tui.Error("unable to load environment", err)
		}
		panic(err)
	}

	client, conn, err := agentClient.NewClient(root)
	if err != nil {
		tui.Error("error connecting to agent", err)
	}
	defer conn.Close()

	_, err = client.ReleaseAccess(context.Background(), &agentGrpc.ReleaseAccessRequest{
		FrontendToken: args[0],
	})
	if err != nil {
		tui.Error("error releasing access", err)
	}

	fmt.Println("success.")
}
