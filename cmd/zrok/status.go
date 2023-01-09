package main

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
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

	zrd, err := zrokdir.Load()
	if err != nil {
		showError("unable to load zrokdir", err)
	}

	if zrd.Env == nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v: Unable to load your local environment!\n\n", warningLabel)
		_, _ = fmt.Fprintf(os.Stderr, "To create a local environment use the %v command.\n", codeStyle.Render("zrok enable"))
	} else {
		_, _ = fmt.Fprintf(os.Stdout, codeStyle.Render("Environment"+":\n"))
		_, _ = fmt.Fprintf(os.Stdout, "\n")

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetStyle(table.StyleColoredDark)
		t.AppendHeader(table.Row{"Property", "Value"})
		t.AppendRow(table.Row{"API Endpoint", zrd.Env.ApiEndpoint})
		t.AppendRow(table.Row{"Secret Token", zrd.Env.Token})
		t.AppendRow(table.Row{"Ziti Identity", zrd.Env.ZId})
		t.Render()
	}

	_, _ = fmt.Fprintf(os.Stderr, "\n")
}
