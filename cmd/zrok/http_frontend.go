package main

import (
	"github.com/michaelquigley/cf"
	"github.com/openziti-test-kitchen/zrok/endpoints/frontend"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	httpCmd.AddCommand(newHttpFrontendCommand().cmd)
}

type httpFrontendCommand struct {
	cmd *cobra.Command
}

func newHttpFrontendCommand() *httpFrontendCommand {
	cmd := &cobra.Command{
		Use:     "frontend [<configPath>]",
		Aliases: []string{"fe"},
		Short:   "Create an HTTP frontend",
		Args:    cobra.RangeArgs(0, 1),
	}
	command := &httpFrontendCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (self *httpFrontendCommand) run(_ *cobra.Command, args []string) {
	cfg := frontend.DefaultConfig()
	if len(args) == 1 {
		if err := cfg.Load(args[0]); err != nil {
			panic(err)
		}
	}
	logrus.Infof(cf.Dump(cfg, cf.DefaultOptions()))
	httpListener, err := frontend.NewHTTP(cfg)
	if err != nil {
		panic(err)
	}
	if err := httpListener.Run(); err != nil {
		panic(err)
	}
}
