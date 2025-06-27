package main

import (
	"fmt"
	"os"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/tui"
	"github.com/spf13/cobra"
)

func init() {
	overviewCmd.AddCommand(newOverviewFrontendsCommand().cmd)
}

type overviewPublicFrontendsCommand struct {
	cmd *cobra.Command
}

func newOverviewFrontendsCommand() *overviewPublicFrontendsCommand {
	cmd := &cobra.Command{
		Use:     "public-frontends",
		Short:   "Show the available public frontends",
		Aliases: []string{"pf"},
		Args:    cobra.NoArgs,
	}
	command := &overviewPublicFrontendsCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *overviewPublicFrontendsCommand) run(_ *cobra.Command, _ []string) {
	root, err := environment.LoadRoot()
	if err != nil {
		if !panicInstead {
			tui.Error("error loading zrokdir", err)
		}
		panic(err)
	}

	if !root.IsEnabled() {
		tui.Error("unable to load environment; did you 'zrok enable'?", nil)
	}

	zrok, err := root.Client()
	if err != nil {
		if !panicInstead {
			tui.Error("error getting zrok client", err)
		}
		panic(err)
	}

	auth := httptransport.APIKeyAuth("X-TOKEN", "header", root.Environment().AccountToken)
	resp, err := zrok.Metadata.ListPublicFrontendsForAccount(nil, auth)
	if err != nil {
		if !panicInstead {
			tui.Error("error listing public frontends", err)
		}
		panic(err)
	}

	if len(resp.Payload.PublicFrontends) > 0 {
		fmt.Println()
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetStyle(table.StyleColoredDark)
		t.AppendHeader(table.Row{"Frontend Name", "URL Template"})
		for _, i := range resp.Payload.PublicFrontends {
			t.AppendRow(table.Row{i.PublicName, i.URLTemplate})
		}
		t.Render()
		fmt.Println()
	} else {
		fmt.Println("no public frontends found")
	}
}
