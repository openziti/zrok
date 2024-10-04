package main

import (
	"github.com/openziti/zrok/tui"
	"github.com/spf13/cobra"
)

func init() {
	agentCmd.AddCommand(newAgentConsoleCommand().cmd)
}

type agentConsoleCommand struct {
	cmd *cobra.Command
}

func newAgentConsoleCommand() *agentConsoleCommand {
	cmd := &cobra.Command{
		Use:   "console",
		Short: "Open the Agent console",
		Args:  cobra.NoArgs,
	}
	command := &agentConsoleCommand{cmd}
	cmd.Run = command.run
	return command
}

func (cmd *agentConsoleCommand) run(_ *cobra.Command, _ []string) {
	if err := openBrowser("http://localhost:8888"); err != nil {
		tui.Error("unable to open agent console at 'http://localhost:8888'", err)
	}
}
