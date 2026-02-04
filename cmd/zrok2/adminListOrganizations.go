package main

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/openziti/zrok/v2/environment"
	"github.com/openziti/zrok/v2/rest_client_zrok/admin"
	"github.com/spf13/cobra"
)

func init() {
	adminListCmd.AddCommand(newAdminListOrganizationsCommand().cmd)
}

type adminListOrganizationsCommand struct {
	cmd *cobra.Command
}

func newAdminListOrganizationsCommand() *adminListOrganizationsCommand {
	cmd := &cobra.Command{
		Use:     "organizations",
		Aliases: []string{"orgs"},
		Short:   "List all organizations",
		Args:    cobra.NoArgs,
	}
	command := &adminListOrganizationsCommand{cmd}
	cmd.Run = command.run
	return command
}

func (c *adminListOrganizationsCommand) run(_ *cobra.Command, _ []string) {
	env, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	zrok, err := env.Client()
	if err != nil {
		panic(err)
	}

	req := admin.NewListOrganizationsParams()
	resp, err := zrok.Admin.ListOrganizations(req, mustGetAdminAuth())
	if err != nil {
		panic(err)
	}

	fmt.Println()
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.AppendHeader(table.Row{"Organization Token", "Description"})
	for _, org := range resp.Payload.Organizations {
		t.AppendRow(table.Row{org.OrganizationToken, org.Description})
	}
	t.Render()
	fmt.Println()
}
