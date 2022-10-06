package main

import (
	"github.com/michaelquigley/cf"
	"github.com/openziti-test-kitchen/zrok/controller"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newGcCmd().cmd)
}

type gcCmd struct {
	cmd *cobra.Command
}

func newGcCmd() *gcCmd {
	cmd := &cobra.Command{
		Use:   "gc <configPath>",
		Short: "Garbage collect a zrok instance",
		Args:  cobra.ExactArgs(1),
	}
	c := &gcCmd{cmd: cmd}
	cmd.Run = c.run
	return c
}

func (gc *gcCmd) run(_ *cobra.Command, args []string) {
	cfg, err := controller.LoadConfig(args[0])
	if err != nil {
		panic(err)
	}
	logrus.Infof(cf.Dump(cfg, cf.DefaultOptions()))
	if err := controller.GC(cfg); err != nil {
		panic(err)
	}
}
