package main

import (
	"github.com/michaelquigley/df/dd"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller"
	"github.com/openziti/zrok/v2/controller/config"
	"github.com/spf13/cobra"
)

func init() {
	adminCmd.AddCommand(newAdminRepairStoreCommand().cmd)
}

type adminRepairStoreCommand struct {
	cmd   *cobra.Command
	apply bool
	batch int
}

func newAdminRepairStoreCommand() *adminRepairStoreCommand {
	cmd := &cobra.Command{
		Use:   "repair-store <configPath>",
		Short: "Release store rows left by incomplete teardowns: environments, shares, accesses, names and mappings (dry run by default)",
		Long: "Report, and with --apply repair, the store rows a teardown would have released had it run, in this order:\n" +
			"live environments of deleted accounts; live shares in deleted environments, released as a teardown\n" +
			"would release them in the store; access frontends of deleted shares; names of deleted accounts;\n" +
			"allocated names no mapping holds; share name mappings to deleted shares (reserved names first, then\n" +
			"allocated names with the names themselves); share name mappings of live shares to deleted names; and\n" +
			"frontend mappings whose share token belongs to no live share. Live shares are expected in the sample\n" +
			"under shares in deleted environments and mappings to deleted names, and nowhere else.\n\n" +
			"Without --apply this is a dry run: nothing is changed and the report counts and samples what would be.\n" +
			"With --apply the rows are repaired in batches of --batch, each its own transaction. Only the store is\n" +
			"touched: no OpenZiti calls (gc collects a released share's objects; the report lists the identities of\n" +
			"released environments for 'zrok2 admin delete identity') and no notifications (dynamic frontends drop\n" +
			"removed mappings at their next reconciliation).",
		Args: cobra.ExactArgs(1),
	}
	command := &adminRepairStoreCommand{cmd: cmd}
	cmd.Flags().BoolVar(&command.apply, "apply", false, "Repair the rows (default is a dry run)")
	cmd.Flags().IntVar(&command.batch, "batch", controller.DefaultRepairStoreBatch, "Rows repaired per transaction")
	cmd.Run = command.run
	return command
}

func (c *adminRepairStoreCommand) run(_ *cobra.Command, args []string) {
	cfg, err := config.LoadConfig(args[0])
	if err != nil {
		panic(err)
	}
	dl.Info(dd.MustInspect(cfg))
	if err := controller.RepairStore(cfg, controller.RepairStoreOptions{Apply: c.apply, Batch: c.batch}); err != nil {
		panic(err)
	}
}
