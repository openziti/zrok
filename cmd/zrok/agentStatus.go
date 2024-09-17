package main

import (
	"context"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/openziti/zrok/agent/agentClient"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/tui"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	agentCmd.AddCommand(newAgentStatusCommand().cmd)
}

type agentStatusCommand struct {
	cmd *cobra.Command
}

func newAgentStatusCommand() *agentStatusCommand {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show the status of the running zrok Agent",
		Args:  cobra.NoArgs,
	}
	command := &agentStatusCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *agentStatusCommand) run(_ *cobra.Command, _ []string) {
	root, err := environment.LoadRoot()
	if err != nil {
		tui.Error("error loading zrokdir", err)
	}

	client, conn, err := agentClient.NewClient(root)
	if err != nil {
		tui.Error("error connecting to agent", err)
	}
	defer conn.Close()

	status, err := client.Status(context.Background(), &agentGrpc.StatusRequest{})
	if err != nil {
		tui.Error("error getting status", err)
	}

	fmt.Println()
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleColoredDark)
	t.AppendHeader(table.Row{"Frontend Token", "Token", "Bind Address"})
	for _, access := range status.GetAccesses() {
		t.AppendRow(table.Row{access.FrontendToken, access.Token, access.BindAddress})
	}
	t.Render()
	fmt.Printf("%d accesses in agent\n", len(status.GetAccesses()))

	fmt.Println()
	t = table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleColoredDark)
	t.AppendHeader(table.Row{"Token", "Reserved", "Share Mode", "Backend Mode", "Target"})
	for _, share := range status.GetShares() {
		t.AppendRow(table.Row{share.Token, share.Reserved, share.ShareMode, share.BackendMode, share.BackendEndpoint})
	}
	t.Render()
	fmt.Printf("%d shares in agent\n", len(status.GetShares()))

	fmt.Println()
}
