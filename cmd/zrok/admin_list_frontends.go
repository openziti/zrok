package main

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok/admin"
	"github.com/openziti-test-kitchen/zrok/zrokdir"
	"github.com/spf13/cobra"
	"os"
	"time"
)

func init() {
	adminListCmd.AddCommand(newAdminListFrontendsCommand().cmd)
}

type adminListFrontendsCommand struct {
	cmd *cobra.Command
}

func newAdminListFrontendsCommand() *adminListFrontendsCommand {
	cmd := &cobra.Command{
		Use:     "frontends",
		Aliases: []string{"fes"},
		Short:   "List global public frontends",
		Args:    cobra.ExactArgs(0),
	}
	command := &adminListFrontendsCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *adminListFrontendsCommand) run(_ *cobra.Command, args []string) {
	zrok, err := zrokdir.ZrokClient(apiEndpoint)
	if err != nil {
		panic(err)
	}

	req := admin.NewListFrontendsParams()
	resp, err := zrok.Admin.ListFrontends(req, mustGetAdminAuth())
	if err != nil {
		panic(err)
	}

	fmt.Println()
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleColoredBright)
	t.AppendHeader(table.Row{"Token", "zId", "Url Template", "Public Name", "Created At", "Updated At"})
	for _, pfe := range resp.Payload {
		t.AppendRow(table.Row{
			pfe.Token,
			pfe.ZID,
			pfe.URLTemplate,
			pfe.PublicName,
			time.UnixMilli(pfe.CreatedAt),
			time.UnixMilli(pfe.UpdatedAt),
		})
	}
	t.Render()
	fmt.Println()
}
