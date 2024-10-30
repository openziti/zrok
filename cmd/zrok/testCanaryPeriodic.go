package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	testCanaryCmd.AddCommand(newTestCanaryPeriodicCommand().cmd)
}

type testCanaryPeriodicCommand struct {
	cmd *cobra.Command
}

func newTestCanaryPeriodicCommand() *testCanaryPeriodicCommand {
	cmd := &cobra.Command{
		Use:   "periodic",
		Short: "Run a periodic canary inspection",
		Args:  cobra.NoArgs,
	}
	command := &testCanaryPeriodicCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (c *testCanaryPeriodicCommand) run(_ *cobra.Command, _ []string) {
	logrus.Info("periodic")
}
