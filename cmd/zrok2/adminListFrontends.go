package main

import (
	"fmt"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/openziti/zrok/v2/environment"
	"github.com/openziti/zrok/v2/rest_client_zrok/admin"
	"github.com/spf13/cobra"
)

func init() {
	adminListCmd.AddCommand(newAdminListFrontendsCommand().cmd)
}

type adminListFrontendsCommand struct {
	cmd   *cobra.Command
	extra bool
}

func newAdminListFrontendsCommand() *adminListFrontendsCommand {
	cmd := &cobra.Command{
		Use:     "frontends",
		Aliases: []string{"fes"},
		Short:   "List global public frontends",
		Args:    cobra.ExactArgs(0),
	}
	command := &adminListFrontendsCommand{cmd: cmd}
	cmd.Flags().BoolVar(&command.extra, "extra", false, "show extra (v1) fields")
	cmd.Run = command.run
	return command
}

func (cmd *adminListFrontendsCommand) run(_ *cobra.Command, _ []string) {
	env, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	zrok, err := env.Client()
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
	t.SetStyle(table.StyleRounded)
	if cmd.extra {
		t.AppendHeader(table.Row{"Frontend Token", "zId", "Public Name", "Url Template", "Permission Mode", "Dynamic", "Created At", "Updated At"})
	} else {
		t.AppendHeader(table.Row{"Frontend Token", "zId", "Public Name", "Permission Mode", "Dynamic", "Updated At"})
	}
	for _, pfe := range resp.Payload {
		if cmd.extra {
			t.AppendRow(table.Row{
				pfe.FrontendToken,
				pfe.ZID,
				pfe.PublicName,
				pfe.URLTemplate,
				pfe.PermissionMode,
				pfe.Dynamic,
				time.UnixMilli(pfe.CreatedAt),
				time.UnixMilli(pfe.UpdatedAt),
			})
		} else {
			t.AppendRow(table.Row{
				pfe.FrontendToken,
				pfe.ZID,
				pfe.PublicName,
				pfe.PermissionMode,
				pfe.Dynamic,
				time.UnixMilli(pfe.UpdatedAt),
			})
		}
	}
	t.Render()
	fmt.Println()
}
