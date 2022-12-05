package main

import (
	"github.com/michaelquigley/cf"
	"github.com/openziti-test-kitchen/zrok/controller"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	adminCmd.AddCommand(newAdminBootstrap().cmd)
}

type adminBootstrap struct {
	cmd *cobra.Command
}

func newAdminBootstrap() *adminBootstrap {
	cmd := &cobra.Command{
		Use:   "bootstrap <configPath>",
		Short: "Bootstrap the underlying Ziti network for zrok",
		Args:  cobra.ExactArgs(1),
	}
	command := &adminBootstrap{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *adminBootstrap) run(_ *cobra.Command, args []string) {
	configPath := args[0]
	inCfg, err := controller.LoadConfig(configPath)
	if err != nil {
		panic(err)
	}
	logrus.Infof(cf.Dump(inCfg, cf.DefaultOptions()))
	if err := controller.Bootstrap(inCfg); err != nil {
		panic(err)
	}
	logrus.Info("bootstrap complete!")
}
