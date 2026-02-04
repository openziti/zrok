package main

import (
	"fmt"
	"github.com/openziti/zrok/v2/build"
	"github.com/openziti/zrok/v2/tui"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newVersionCommand().cmd)
}

type versionCommand struct {
	cmd *cobra.Command
}

func newVersionCommand() *versionCommand {
	cmd := &cobra.Command{
		Use:  "version",
		Args: cobra.ExactArgs(0),
	}
	command := &versionCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *versionCommand) run(_ *cobra.Command, _ []string) {
	fmt.Println(tui.SeriousBusiness.Render("               _    \n _____ __ ___ | | __\n|_  / '__/ _ \\| |/ /\n / /| | | (_) |   < \n/___|_|  \\___/|_|\\_\\") + "\n\n" + tui.Attention.Render(build.String()) + "\n")
}
