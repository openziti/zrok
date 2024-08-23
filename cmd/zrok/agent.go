package main

import (
	"github.com/openziti/zrok/agent"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/tui"
	"github.com/spf13/cobra"
)

func init() {
	agentCmd.AddCommand(newAgentStartCommand().cmd)
}

type agentStartCommand struct {
	cmd *cobra.Command
}

func newAgentStartCommand() *agentStartCommand {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Launch a zrok agent",
		Args:  cobra.NoArgs,
	}
	command := &agentStartCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *agentStartCommand) run(_ *cobra.Command, _ []string) {
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
