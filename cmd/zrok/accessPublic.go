package main

import (
	"fmt"
	"github.com/michaelquigley/cf"
	"github.com/openziti/zrok/endpoints/publicProxy"
	"github.com/openziti/zrok/tui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var accessPublicCmd *accessPublicCommand

func init() {
	accessPublicCmd = newAccessPublicCommand()
	accessCmd.AddCommand(accessPublicCmd.cmd)
}

type accessPublicCommand struct {
	cmd *cobra.Command
}

func newAccessPublicCommand() *accessPublicCommand {
	cmd := &cobra.Command{
		Use:     "public [<configPath>]",
		Aliases: []string{"fe"},
		Short:   "Create a public access HTTP frontend",
		Args:    cobra.RangeArgs(0, 1),
	}
	command := &accessPublicCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *accessPublicCommand) run(_ *cobra.Command, args []string) {
	cfg := publicProxy.DefaultConfig()
	if len(args) == 1 {
		if err := cfg.Load(args[0]); err != nil {
			if !panicInstead {
				tui.Error(fmt.Sprintf("unable to load configuration '%v'", args[0]), err)
			}
			panic(err)
		}
	}
	logrus.Infof(cf.Dump(cfg, cf.DefaultOptions()))
	frontend, err := publicProxy.NewHTTP(cfg)
	if err != nil {
		if !panicInstead {
			tui.Error("unable to create http frontend", err)
		}
		panic(err)
	}
	if err := frontend.Run(); err != nil {
		if !panicInstead {
			tui.Error("unable to run http frontend", err)
		}
		panic(err)
	}
}
