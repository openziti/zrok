package main

import (
	"github.com/michaelquigley/df/dd"
	"github.com/openziti/zrok/controller"
	"github.com/openziti/zrok/controller/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	adminCmd.AddCommand(newAdminUnbootstrap().cmd)
}

type adminUnbootstrap struct {
	cmd *cobra.Command
}

func newAdminUnbootstrap() *adminUnbootstrap {
	cmd := &cobra.Command{
		Use:   "unbootstrap <configPath>",
		Short: "Unbootstrap the underlying Ziti network from zrok",
		Args:  cobra.ExactArgs(1),
	}
	command := &adminUnbootstrap{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *adminUnbootstrap) run(_ *cobra.Command, args []string) {
	cfg, err := config.LoadConfig(args[0])
	if err != nil {
		panic(err)
	}
	logrus.Info(dd.MustInspect(cfg))
	if err := controller.Unbootstrap(cfg); err != nil {
		panic(err)
	}
	logrus.Info("unbootstrap complete!")
}
