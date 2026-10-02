package main

import (
	"time"

	"github.com/michaelquigley/df/dd"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller"
	"github.com/openziti/zrok/v2/controller/config"
	"github.com/spf13/cobra"
)

func init() {
	adminCmd.AddCommand(newAdminGcCommand().cmd)
}

type adminGcCommand struct {
	cmd    *cobra.Command
	delete bool
	minAge time.Duration
}

func newAdminGcCommand() *adminGcCommand {
	cmd := &cobra.Command{
		Use:   "gc <configPath>",
		Short: "Garbage collect orphaned OpenZiti objects (dry run by default)",
		Long: "Report, and with --delete remove, the OpenZiti services, configs, service policies and service edge\n" +
			"router policies whose zrokShareToken tag names no live share.\n\n" +
			"Without --delete this is a dry run: nothing is deleted and the report lists what would be.\n" +
			"Objects without a zrokShareToken tag are never touched, and orphans younger than --min-age\n" +
			"(default 24h) are skipped so that a share being created is not collected.",
		Args: cobra.ExactArgs(1),
	}
	command := &adminGcCommand{cmd: cmd}
	cmd.Flags().BoolVar(&command.delete, "delete", false, "Delete the orphaned objects (default is a dry run)")
	cmd.Flags().DurationVar(&command.minAge, "min-age", controller.DefaultGCMinAge, "Skip orphaned objects younger than this")
	cmd.Run = command.run
	return command
}

func (gc *adminGcCommand) run(_ *cobra.Command, args []string) {
	cfg, err := config.LoadConfig(args[0])
	if err != nil {
		panic(err)
	}
	dl.Info(dd.MustInspect(cfg))
	if err := controller.GC(cfg, controller.GCOptions{Delete: gc.delete, MinAge: gc.minAge}); err != nil {
		panic(err)
	}
}
