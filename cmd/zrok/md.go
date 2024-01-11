package main

import (
	"fmt"
	"github.com/openziti/zrok/drives/sync"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/tui"
	"github.com/spf13/cobra"
	"net/url"
)

func init() {
	rootCmd.AddCommand(newMdCommand().cmd)
}

type mdCommand struct {
	cmd *cobra.Command
}

func newMdCommand() *mdCommand {
	cmd := &cobra.Command{
		Use:     "md <target>",
		Short:   "Make directory at <target> ('http://', 'zrok://', 'file://')",
		Aliases: []string{"mkdir"},
		Args:    cobra.ExactArgs(1),
	}
	command := &mdCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *mdCommand) run(_ *cobra.Command, args []string) {
	targetUrl, err := url.Parse(args[0])
	if err != nil {
		tui.Error(fmt.Sprintf("invalid target '%v'", args[0]), err)
	}
	if targetUrl.Scheme == "" {
		targetUrl.Scheme = "file"
	}

	root, err := environment.LoadRoot()
	if err != nil {
		tui.Error("error loading root", err)
	}

	target, err := sync.TargetForURL(targetUrl, root)
	if err != nil {
		tui.Error(fmt.Sprintf("error creating target for '%v'", targetUrl), err)
	}

	if err := target.Mkdir("/"); err != nil {
		tui.Error("error creating directory", err)
	}
}
