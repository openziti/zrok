//go:build windows

package main

import "github.com/spf13/cobra"

func init() {
	agentServiceCmd.AddCommand(newAgentServiceStartCommand().cmd)
}

type agentServiceStartCommand struct {
	cmd *cobra.Command
}

func newAgentServiceStartCommand() *agentServiceStartCommand {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the agent as a service (on Windows)",
	}
	out := &agentServiceStartCommand{cmd: cmd}
	cmd.Run = out.run
	return out
}

func (cmd *agentServiceStartCommand) run(_ *cobra.Command, _ []string) {

}
