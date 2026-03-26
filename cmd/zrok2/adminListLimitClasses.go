package main

import (
	"fmt"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/openziti/zrok/v2/environment"
	"github.com/openziti/zrok/v2/rest_client_zrok/admin"
	"github.com/openziti/zrok/v2/util"
	"github.com/spf13/cobra"
)

func init() {
	adminListCmd.AddCommand(newAdminListLimitClassesCommand().cmd)
}

type adminListLimitClassesCommand struct {
	cmd *cobra.Command
}

func newAdminListLimitClassesCommand() *adminListLimitClassesCommand {
	cmd := &cobra.Command{
		Use:     "limit-classes <label>",
		Aliases: []string{"lcs"},
		Short:   "List limit classes matching the specified label",
		Args:    cobra.ExactArgs(1),
	}
	command := &adminListLimitClassesCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *adminListLimitClassesCommand) run(_ *cobra.Command, args []string) {
	env, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	zrok, err := env.Client()
	if err != nil {
		panic(err)
	}

	req := admin.NewListLimitClassesParams()
	req.Body.Label = args[0]

	resp, err := zrok.Admin.ListLimitClasses(req, mustGetAdminAuth())
	if err != nil {
		panic(err)
	}

	fmt.Println()
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.AppendHeader(table.Row{"ID", "Label", "Backend Mode", "Envs", "Shares", "Reserved", "Unique Names", "Share FEs", "Period Min", "Rx", "Tx", "Total", "Action", "Updated At"})
	for _, lc := range resp.Payload {
		t.AppendRow(table.Row{
			lc.ID,
			lc.Label,
			lc.BackendMode,
			lc.Environments,
			lc.Shares,
			lc.ReservedShares,
			lc.UniqueNames,
			lc.ShareFrontends,
			lc.PeriodMinutes,
			util.BytesToSize(lc.RxBytes),
			util.BytesToSize(lc.TxBytes),
			util.BytesToSize(lc.TotalBytes),
			lc.LimitAction,
			time.UnixMilli(lc.UpdatedAt),
		})
	}
	t.Render()
	fmt.Println()
}
