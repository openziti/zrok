package main

import (
	"fmt"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/rest_client_zrok/metadata"
	"github.com/openziti/zrok/tui"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	organizationAdminCmd.AddCommand(newOrgAdminListCommand().cmd)
}

type orgAdminListCommand struct {
	cmd *cobra.Command
}

func newOrgAdminListCommand() *orgAdminListCommand {
	cmd := &cobra.Command{
		Use:   "list <organizationToken>",
		Short: "List the members of an organization",
		Args:  cobra.ExactArgs(1),
	}
	command := &orgAdminListCommand{cmd}
	cmd.Run = command.run
	return command
}

func (c *orgAdminListCommand) run(_ *cobra.Command, args []string) {
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

	req := metadata.NewListOrgMembersParams()
	req.OrganizationToken = args[0]

	resp, err := zrok.Metadata.ListOrgMembers(req, auth)
	if err != nil {
		if !panicInstead {
			tui.Error("error listing organization members", err)
		}
		panic(err)
	}

	fmt.Println()
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleColoredDark)
	t.AppendHeader(table.Row{"Email", "Admin?"})
	for _, member := range resp.Payload.Members {
		t.AppendRow(table.Row{member.Email, member.Admin})
	}
	t.Render()
	fmt.Println()
}
