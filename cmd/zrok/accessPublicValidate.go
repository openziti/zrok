package main

import (
	"fmt"

	"github.com/michaelquigley/df"
	"github.com/openziti/zrok/endpoints/publicProxy"
	"github.com/openziti/zrok/tui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	accessPublicCmd.cmd.AddCommand(newAccessPublicValidateCommand().cmd)
}

type accessPublicValidateCommand struct {
	cmd *cobra.Command
}

func newAccessPublicValidateCommand() *accessPublicValidateCommand {
	cmd := &cobra.Command{
		Use:   "validate <configPath>",
		Short: "Validate a zrok access public configuration document",
		Args:  cobra.ExactArgs(1),
	}
	command := &accessPublicValidateCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *accessPublicValidateCommand) run(_ *cobra.Command, args []string) {
	cfg := publicProxy.DefaultConfig()
	if err := cfg.Load(args[0]); err != nil {
		tui.Error(fmt.Sprintf("unable to load configuration '%v'", args[0]), err)
	}
	logrus.Info(df.Inspect(cfg))
}
