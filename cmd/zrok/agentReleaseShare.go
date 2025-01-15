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
	agentReleaseCmd.AddCommand(newAgentReleaseShareCommand().cmd)
}

type agentReleaseShareCommand struct {
	cmd *cobra.Command
}

func newAgentReleaseShareCommand() *agentReleaseShareCommand {
	cmd := &cobra.Command{
		Use:   "share <token>",
		Short: "Release a share from the zrok Agent",
		Args:  cobra.ExactArgs(1),
	}
	command := &agentReleaseShareCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *agentReleaseShareCommand) run(_ *cobra.Command, args []string) {
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

	_, err = client.ReleaseShare(context.Background(), &agentGrpc.ReleaseShareRequest{
		Token: args[0],
	})
	if err != nil {
		tui.Error("error releasing share", err)
	}

	fmt.Println("success.")
}
