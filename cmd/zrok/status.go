package main

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/tui"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	rootCmd.AddCommand(newStatusCommand().cmd)
}

type statusCommand struct {
	secrets bool
	cmd     *cobra.Command
}

func newStatusCommand() *statusCommand {
	cmd := &cobra.Command{
		Use:     "status",
		Short:   "Show the current environment status",
		Aliases: []string{"st"},
		Args:    cobra.ExactArgs(0),
	}
	command := &statusCommand{cmd: cmd}
	cmd.Flags().BoolVar(&command.secrets, "secrets", false, "Show secrets in status output")
	cmd.Run = command.run
	return command
}

func (cmd *statusCommand) run(_ *cobra.Command, _ []string) {
	_, _ = fmt.Fprintf(os.Stderr, "\n")

	zrd, err := environment.Load()
	if err != nil {
		tui.Error("error loading environment", err)
	}

	_, _ = fmt.Fprintf(os.Stdout, tui.Code.Render("Config")+":\n\n")
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleColoredDark)
	t.AppendHeader(table.Row{"Config", "Value", "Source"})
	apiEndpoint, from := zrd.ApiEndpoint()
	t.AppendRow(table.Row{"apiEndpoint", apiEndpoint, from})
	t.Render()
	_, _ = fmt.Fprintf(os.Stderr, "\n")

	if zrd.Env == nil {
		tui.Warning("Unable to load your local environment!\n")
		_, _ = fmt.Fprintf(os.Stderr, "To create a local environment use the %v command.\n", tui.Code.Render("zrok enable"))
	} else {
		_, _ = fmt.Fprintf(os.Stdout, tui.Code.Render("Environment")+":\n\n")

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetStyle(table.StyleColoredDark)
		t.AppendHeader(table.Row{"Property", "Value"})
		if cmd.secrets {
			t.AppendRow(table.Row{"Secret Token", zrd.Env.Token})
			t.AppendRow(table.Row{"Ziti Identity", zrd.Env.ZId})
		} else {
			secretToken := "<<SET>>"
			if zrd.Env.Token == "" {
				secretToken = "<<UNSET>>"
			}
			t.AppendRow(table.Row{"Secret Token", secretToken})

			zId := "<<SET>>"
			if zrd.Env.ZId == "" {
				zId = "<<UNSET>>"
			}
			t.AppendRow(table.Row{"Ziti Identity", zId})
		}
		t.Render()
	}
	_, _ = fmt.Fprintf(os.Stdout, "\n")
}
