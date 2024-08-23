package main

import (
	"github.com/openziti/zrok/agent"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/tui"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newAgentCommand().cmd)
}

type agentCommand struct {
	cmd *cobra.Command
}

func newAgentCommand() *agentCommand {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Launch a zrok agent",
		Args:  cobra.NoArgs,
	}
	command := &agentCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *agentCommand) run(_ *cobra.Command, _ []string) {
	root, err := environment.LoadRoot()
	if err != nil {
		tui.Error("error loading zrokdir", err)
	}

	if !root.IsEnabled() {
		tui.Error("unable to load environment; did you 'zrok enable'?", nil)
	}

	a, err := agent.NewAgent(root)
	if err != nil {
		tui.Error("error creating agent", err)
	}

	if err := a.Run(); err != nil {
		tui.Error("agent aborted", err)
	}
}
