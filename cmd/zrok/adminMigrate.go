package main

import (
	"github.com/michaelquigley/cf"
	"github.com/openziti/zrok/controller/config"
	"github.com/openziti/zrok/controller/store"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	adminCmd.AddCommand(newAdminMigrate().cmd)
}

type adminMigrate struct {
	cmd *cobra.Command
}

func newAdminMigrate() *adminMigrate {
	cmd := &cobra.Command{
		Use:   "migrate <configPath>",
		Short: "Migrate the underlying datastore",
		Args:  cobra.ExactArgs(1),
	}
	command := &adminMigrate{cmd}
	cmd.Run = command.run
	return command
}

func (cmd *adminMigrate) run(_ *cobra.Command, args []string) {
	configPath := args[0]
	inCfg, err := config.LoadConfig(configPath)
	if err != nil {
		panic(err)
	}

	logrus.Info(cf.Dump(inCfg, cf.DefaultOptions()))

	// override the 'disable_auto_migration' setting... the user is requesting a migration here.
	inCfg.Store.DisableAutoMigration = false

	if _, err := store.Open(inCfg.Store); err != nil {
		panic(err)
	}
	logrus.Info("migration complete")
}
