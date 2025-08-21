package main

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/rest_client_zrok/admin"
	"github.com/spf13/cobra"
	"os"
	"time"
)

func init() {
	adminListCmd.AddCommand(newAdminListNamespacesCommand().cmd)
}

type adminListNamespacesCommand struct {
	cmd *cobra.Command
}

func newAdminListNamespacesCommand() *adminListNamespacesCommand {
	cmd := &cobra.Command{
		Use:   "namespaces",
		Short: "List all namespaces",
		Args:  cobra.NoArgs,
	}
	command := &adminListNamespacesCommand{cmd}
	cmd.Run = command.run
	return command
}

func (c *adminListNamespacesCommand) run(_ *cobra.Command, _ []string) {
	env, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	zrok, err := env.Client()
	if err != nil {
		panic(err)
	}

	req := admin.NewListNamespacesParams()
	resp, err := zrok.Admin.ListNamespaces(req, mustGetAdminAuth())
	if err != nil {
		panic(err)
	}

	fmt.Println()
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleColoredDark)
	t.AppendHeader(table.Row{"Namespace Token", "Name", "Description", "Open", "Created", "Updated"})
	for _, ns := range resp.Payload {
		created := time.Unix(ns.CreatedAt, 0).Format("2006-01-02 15:04:05")
		updated := time.Unix(ns.UpdatedAt, 0).Format("2006-01-02 15:04:05")
		t.AppendRow(table.Row{ns.NamespaceToken, ns.Name, ns.Description, ns.Open, created, updated})
	}
	t.Render()
	fmt.Println()
}