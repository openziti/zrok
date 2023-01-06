package main

import (
	"fmt"
	"github.com/openziti-test-kitchen/zrok/zrokdir"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	rootCmd.AddCommand(newStatusCommand().cmd)
}

type statusCommand struct {
	cmd *cobra.Command
}

func newStatusCommand() *statusCommand {
	cmd := &cobra.Command{
		Use:     "status",
		Short:   "Show the current environment status",
		Aliases: []string{"st"},
		Args:    cobra.ExactArgs(0),
	}
	command := &statusCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *statusCommand) run(_ *cobra.Command, _ []string) {
	_, _ = fmt.Fprintf(os.Stderr, "\n")

	_, err := zrokdir.LoadEnvironment()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v: Unable to load your local environment! (%v)\n\n", warningLabel, err)
		_, _ = fmt.Fprintf(os.Stderr, "To create a local environment use the %v command.\n", codeStyle.Render("zrok enable"))
	}

	_, _ = fmt.Fprintf(os.Stderr, "\n")
}
