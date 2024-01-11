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
	rootCmd.AddCommand(newMvCommand().cmd)
}

type mvCommand struct {
	cmd *cobra.Command
}

func newMvCommand() *mvCommand {
	cmd := &cobra.Command{
		Use:     "mv <target> <newPath>",
		Short:   "Move the drive <target> to <newPath> ('http://', 'zrok://', 'file://')",
		Aliases: []string{"move"},
		Args:    cobra.ExactArgs(2),
	}
	command := &mvCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *mvCommand) run(_ *cobra.Command, args []string) {
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

	if err := target.Move("/", args[1]); err != nil {
		tui.Error("error moving", err)
	}
}
