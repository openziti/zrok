package main

import (
	"github.com/michaelquigley/cf"
	"github.com/openziti/zrok/controller"
	"github.com/openziti/zrok/controller/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	adminCmd.AddCommand(newAdminBootstrap().cmd)
}

type adminBootstrap struct {
	cmd          *cobra.Command
	skipCtrl     bool
	skipFrontend bool
}

func newAdminBootstrap() *adminBootstrap {
	cmd := &cobra.Command{
		Use:   "bootstrap <configPath>",
		Short: "Bootstrap the underlying Ziti network for zrok",
		Args:  cobra.ExactArgs(1),
	}
	command := &adminBootstrap{cmd: cmd}
	cmd.Run = command.run
	cmd.Flags().BoolVar(&command.skipCtrl, "skip-ctrl", false, "Skip controller (ctrl) identity bootstrapping")
	cmd.Flags().BoolVar(&command.skipFrontend, "skip-frontend", false, "Slip frontend identity bootstrapping")
	return command
}

func (cmd *adminBootstrap) run(_ *cobra.Command, args []string) {
	configPath := args[0]
	inCfg, err := config.LoadConfig(configPath)
	if err != nil {
		panic(err)
	}
	logrus.Infof(cf.Dump(inCfg, cf.DefaultOptions()))
	if err := controller.Bootstrap(cmd.skipCtrl, cmd.skipFrontend, inCfg); err != nil {
		panic(err)
	}
	logrus.Info("bootstrap complete!")
}
