package main

import (
	"fmt"
	"github.com/openziti/zrok/environment/env_v0_3"
	"github.com/openziti/zrok/tui"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newConsoleCommand().cmd)
}

type consoleCommand struct {
	cmd *cobra.Command
}

func newConsoleCommand() *consoleCommand {
	cmd := &cobra.Command{
		Use:   "console",
		Short: "Open the web console",
		Args:  cobra.ExactArgs(0),
	}
	command := &consoleCommand{cmd}
	cmd.Run = command.run
	return command
}

func (cmd *consoleCommand) run(_ *cobra.Command, _ []string) {
	env, err := env_v0_3.Load()
	if err != nil {
		tui.Error("unable to load environment", err)
	}

	apiEndpoint, _ := env.ApiEndpoint()
	if err := openBrowser(apiEndpoint); err != nil {
		tui.Error(fmt.Sprintf("unable to open '%v'", apiEndpoint), err)
	}
}
