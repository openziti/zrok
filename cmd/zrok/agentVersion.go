package main

import (
	"context"
	grpc2 "github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/tui"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	agentSocket, err := root.AgentSocket()
	if err != nil {
		tui.Error("error getting agent socket", err)
	}

	conn, err := grpc.NewClient("unix://"+agentSocket, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		tui.Error("error connecting to agent socket", err)
	}
	defer conn.Close()
	client := grpc2.NewAgentClient(conn)

	v, err := client.Version(context.Background(), &grpc2.VersionRequest{})
	if err != nil {
		tui.Error("error getting agent version", err)
	}

	println(*v.V)
}
