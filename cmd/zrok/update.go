package main

import (
	"fmt"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/tui"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newUpdateCommand().cmd)
}

type updateCommand struct {
	cmd *cobra.Command
}

func newUpdateCommand() *updateCommand {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update your environment to the latest version",
		Args:  cobra.NoArgs,
	}
	command := &updateCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *updateCommand) run(_ *cobra.Command, _ []string) {
	r, err := environment.LoadRoot()
	if err != nil {
		if !panicInstead {
			tui.Error("unable to load environment", err)
		}
		panic(err)
	}

	if environment.IsLatest(r) {
		fmt.Printf("zrok environment is already the latest version at '%v'\n", r.Metadata().V)
		return
	}

	newR, err := environment.UpdateRoot(r)
	if err != nil {
		if !panicInstead {
			tui.Error("unable to update environment", err)
		}
		panic(err)
	}

	fmt.Printf("environment updated to '%v'\n", newR.Metadata().V)
}
