package main

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/rest_client_zrok/admin"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	adminListCmd.AddCommand(newAdminListOrgMembersCommand().cmd)
}

type adminListOrgMembersCommand struct {
	cmd *cobra.Command
}

func newAdminListOrgMembersCommand() *adminListOrgMembersCommand {
	cmd := &cobra.Command{
		Use:     "org-members <organizationToken>",
		Aliases: []string{"members"},
		Short:   "List the members of the specified organization",
		Args:    cobra.ExactArgs(1),
	}
	command := &adminListOrgMembersCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *adminListOrgMembersCommand) run(_ *cobra.Command, args []string) {
	env, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	zrok, err := env.Client()
	if err != nil {
		panic(err)
	}

	req := admin.NewListOrganizationMembersParams()
	req.Body.OrganizationToken = args[0]

	resp, err := zrok.Admin.ListOrganizationMembers(req, mustGetAdminAuth())
	if err != nil {
		panic(err)
	}

	fmt.Println()
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleColoredDark)
	t.AppendHeader(table.Row{"Account Email", "Admin?"})
	for _, member := range resp.Payload.Members {
		t.AppendRow(table.Row{member.Email, member.Admin})
	}
	t.Render()
	fmt.Println()
}
