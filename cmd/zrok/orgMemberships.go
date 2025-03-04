package main

import (
	"fmt"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/tui"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	organizationCmd.AddCommand(newOrgMembershipsCommand().cmd)
}

type orgMembershipsCommand struct {
	cmd *cobra.Command
}

func newOrgMembershipsCommand() *orgMembershipsCommand {
	cmd := &cobra.Command{
		Use:   "memberships",
		Short: "List the organization memberships for my account",
		Args:  cobra.NoArgs,
	}
	command := &orgMembershipsCommand{cmd}
	cmd.Run = command.run
	return command
}

func (c *orgMembershipsCommand) run(_ *cobra.Command, _ []string) {
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
			tui.Error("error loading zrokdir", err)
		}
		panic(err)
	}
	auth := httptransport.APIKeyAuth("X-TOKEN", "header", root.Environment().AccountToken)

	in, err := zrok.Metadata.ListMemberships(nil, auth)
	if err != nil {
		if !panicInstead {
			tui.Error("error listing memberships", err)
		}
		panic(err)
	}

	if len(in.Payload.Memberships) > 0 {
		fmt.Println()
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetStyle(table.StyleColoredDark)
		t.AppendHeader(table.Row{"Organization Token", "Description", "Admin?"})
		for _, i := range in.Payload.Memberships {
			t.AppendRow(table.Row{i.OrganizationToken, i.Description, i.Admin})
		}
		t.Render()
		fmt.Println()
	} else {
		fmt.Println("no organization memberships.")
	}
}
