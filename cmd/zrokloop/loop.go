package main

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(newRun().cmd)
}

type run struct {
	cmd *cobra.Command
}

func newRun() *run {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Start a loop agent",
		Args:  cobra.ExactArgs(0),
	}
	r := &run{cmd: cmd}
	cmd.Run = r.run
	return r
}

func (r *run) run(_ *cobra.Command, _ []string) {
}
