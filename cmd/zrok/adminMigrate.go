package main

import (
	"github.com/michaelquigley/df/dd"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/controller/config"
	"github.com/openziti/zrok/controller/store"
	"github.com/spf13/cobra"
)

func init() {
	adminCmd.AddCommand(newAdminMigrate().cmd)
}

type adminMigrate struct {
	cmd   *cobra.Command
	steps int
}

func newAdminMigrate() *adminMigrate {
	cmd := &cobra.Command{
		Use:   "migrate <configPath>",
		Short: "Migrate the underlying datastore",
		Args:  cobra.ExactArgs(1),
	}
	command := &adminMigrate{cmd: cmd}
	cmd.Flags().IntVar(&command.steps, "down", 0, "migrate down N steps (0 = migrate up)")
	cmd.Run = command.run
	return command
}

func (cmd *adminMigrate) run(_ *cobra.Command, args []string) {
	configPath := args[0]
	inCfg, err := config.LoadConfig(configPath)
	if err != nil {
		panic(err)
	}

	dl.Info(dd.MustInspect(inCfg))

	// disable auto-migration, we'll control it manually
	inCfg.Store.DisableAutoMigration = true

	str, err := store.Open(inCfg.Store)
	if err != nil {
		panic(err)
	}
	defer str.Close()

	if cmd.steps > 0 {
		if err := str.MigrateDown(inCfg.Store, cmd.steps); err != nil {
			panic(err)
		}
		dl.Infof("migrated down %d steps", cmd.steps)
	} else {
		// default behavior - migrate up
		inCfg.Store.DisableAutoMigration = false
		if _, err := store.Open(inCfg.Store); err != nil {
			panic(err)
		}
		dl.Info("migration complete")
	}
}
