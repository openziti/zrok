package main

import (
	"fmt"
	"github.com/michaelquigley/cf"
	"github.com/openziti-test-kitchen/zrok/endpoints/public_frontend"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	accessCmd.AddCommand(newAccessPublicCommand().cmd)
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

func (self *accessPublicCommand) run(_ *cobra.Command, args []string) {
	cfg := public_frontend.DefaultConfig()
	if len(args) == 1 {
		if err := cfg.Load(args[0]); err != nil {
			if !panicInstead {
				showError(fmt.Sprintf("unable to load configuration '%v'", args[0]), err)
			}
			panic(err)
		}
	}
	logrus.Infof(cf.Dump(cfg, cf.DefaultOptions()))
	frontend, err := public_frontend.NewHTTP(cfg)
	if err != nil {
		if !panicInstead {
			showError("unable to create http frontend", err)
		}
		panic(err)
	}
	if err := frontend.Run(); err != nil {
		if !panicInstead {
			showError("unable to run http frontend", err)
		}
		panic(err)
	}
}
