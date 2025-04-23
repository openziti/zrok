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
	agentCmd.AddCommand(newAgentConsoleCommand().cmd)
}

type agentConsoleCommand struct {
	cmd      *cobra.Command
	headless bool
}

func newAgentConsoleCommand() *agentConsoleCommand {
	cmd := &cobra.Command{
		Use:   "console",
		Short: "Open the Agent console",
		Args:  cobra.NoArgs,
	}
	command := &agentConsoleCommand{cmd: cmd}
	cmd.Flags().BoolVar(&command.headless, "headless", false, "Do not attempt to open console, just emit console URL")
	cmd.Run = command.run
	return command
}

func (cmd *agentConsoleCommand) run(_ *cobra.Command, _ []string) {
	root, err := environment.LoadRoot()
	if err != nil {
		tui.Error("error loading zrokdir", err)
	}

	client, conn, err := agentClient.NewClient(root)
	if err != nil {
		tui.Error("error connecting to agent", err)
	}
	defer func() { _ = conn.Close() }()

	v, err := client.Version(context.Background(), &agentGrpc.VersionRequest{})
	if err != nil {
		tui.Error("error getting agent version", err)
	}

	if cmd.headless {
		fmt.Println("http://" + v.ConsoleEndpoint)
	} else {
		fmt.Printf("opening default web browser for: http://%v\n", v.ConsoleEndpoint)
		if err := openBrowser("http://" + v.ConsoleEndpoint); err != nil {
			tui.Error(fmt.Sprintf("unable to open agent console at 'http://%v'", v.ConsoleEndpoint), err)
		}
	}
}
