//go:build windows

package main

import "github.com/spf13/cobra"

func init() {
	agentCmd.AddCommand(agentServiceCmd)
}

const agentServiceName = "zrok"

var agentServiceCmd = &cobra.Command{
	Use:     "service",
	Short:   "Command for managing the agent as a service (on Windows)",
	Aliases: []string{"svc"},
}
