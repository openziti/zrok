package main

import (
	"fmt"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/sdk"
	"github.com/openziti/zrok/tui"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newOverviewCommand().cmd)
}

type overviewCommand struct {
	cmd *cobra.Command
}

func newOverviewCommand() *overviewCommand {
	cmd := &cobra.Command{
		Use:   "overview",
		Short: "Retrieve all of the zrok account details (environments, shares) as JSON",
		Args:  cobra.ExactArgs(0),
	}
	command := &overviewCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *overviewCommand) run(_ *cobra.Command, _ []string) {
	root, err := environment.LoadRoot()
	if err != nil {
		if !panicInstead {
			tui.Error("error loading zrokdir", err)
		}
		panic(err)
	}

	if !root.IsEnabled() {
		tui.Error("unable to load environment; did you 'zrok enable'?", nil)
	}

	json, err := sdk.Overview(root)
	if err != nil {
		if !panicInstead {
			tui.Error("error loading zrokdir", err)
		}
		panic(err)
	}

	fmt.Println(json)
}
