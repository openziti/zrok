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

	env, err := environment.LoadRoot()
	if err != nil {
		tui.Error("error loading environment", err)
	}

	if !environment.IsLatest(env) {
		tui.Warning(fmt.Sprintf("Your environment is out of date ('%v'), use '%v' to update (make a backup before updating!)\n", env.Metadata().V, tui.Code.Render("zrok update")))
	}

	_, _ = fmt.Fprintf(os.Stdout, tui.Code.Render("Config")+":\n\n")
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleColoredDark)
	t.AppendHeader(table.Row{"Config", "Value", "Source"})
	apiEndpoint, apiEndpointFrom := env.ApiEndpoint()
	t.AppendRow(table.Row{"apiEndpoint", apiEndpoint, apiEndpointFrom})
	defaultFrontend, defaultFrontendFrom := env.DefaultFrontend()
	t.AppendRow(table.Row{"defaultFrontend", defaultFrontend, defaultFrontendFrom})
	headless, headlessFrom := env.Headless()
	t.AppendRow(table.Row{"headless", headless, headlessFrom})
	t.Render()
	_, _ = fmt.Fprintf(os.Stderr, "\n")

	if !env.IsEnabled() {
		_, _ = fmt.Fprintf(os.Stderr, "To create a local environment use the %v command.\n", tui.Code.Render("zrok enable"))
	} else {
		_, _ = fmt.Fprintf(os.Stdout, tui.Code.Render("Environment")+":\n\n")

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetStyle(table.StyleColoredDark)
		t.AppendHeader(table.Row{"Property", "Value"})
		if cmd.secrets {
			t.AppendRow(table.Row{"Account Token", env.Environment().AccountToken})
			t.AppendRow(table.Row{"Ziti Identity", env.Environment().ZitiIdentity})
		} else {
			secretToken := "<<SET>>"
			if env.Environment().AccountToken == "" {
				secretToken = "<<UNSET>>"
			}
			t.AppendRow(table.Row{"Account Token", secretToken})

			zId := "<<SET>>"
			if env.Environment().ZitiIdentity == "" {
				zId = "<<UNSET>>"
			}
			t.AppendRow(table.Row{"Ziti Identity", zId})
		}
		t.Render()
	}
	_, _ = fmt.Fprintf(os.Stdout, "\n")
}
