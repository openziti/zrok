package main

import (
	"context"
	"github.com/openziti/zrok/agent/agentClient"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/tui"
	"github.com/spf13/cobra"
)

func init() {
	agentCmd.AddCommand(newAgentVersionCommand().cmd)
}

type agentVersionCommand struct {
	cmd *cobra.Command
}

func newAgentVersionCommand() *agentVersionCommand {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Retrieve the running zrok Agent version",
		Args:  cobra.NoArgs,
	}
	command := &agentVersionCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *agentVersionCommand) run(_ *cobra.Command, _ []string) {
	root, err := environment.LoadRoot()
	if err != nil {
		tui.Error("error loading zrokdir", err)
	}

	client, conn, err := agentClient.NewClient(root)
	if err != nil {
		tui.Error("error getting agent client", err)
	}
	defer conn.Close()

	v, err := client.Version(context.Background(), &agentGrpc.VersionRequest{})
	if err != nil {
		tui.Error("error getting agent version", err)
	}

	println(v.GetV())
}
