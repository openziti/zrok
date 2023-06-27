package main

import (
	"github.com/michaelquigley/cf"
	"github.com/openziti/zrok/controller"
	"github.com/openziti/zrok/controller/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	adminCmd.AddCommand(newAdminGcCommand().cmd)
}

type adminGcCommand struct {
	cmd *cobra.Command
}

func newAdminGcCommand() *adminGcCommand {
	cmd := &cobra.Command{
		Use:   "gc <configPath>",
		Short: "Garbage collect a zrok instance",
		Args:  cobra.ExactArgs(1),
	}
	command := &adminGcCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (gc *adminGcCommand) run(_ *cobra.Command, args []string) {
	cfg, err := config.LoadConfig(args[0])
	if err != nil {
		panic(err)
	}
	logrus.Infof(cf.Dump(cfg, cf.DefaultOptions()))
	if err := controller.GC(cfg); err != nil {
		panic(err)
	}
}
