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
	rootCmd.AddCommand(newRmCommand().cmd)
}

type rmCommand struct {
	cmd *cobra.Command
}

func newRmCommand() *rmCommand {
	cmd := &cobra.Command{
		Use:     "rm <target>",
		Short:   "Remove (delete) the contents of drive <target> ('http://', 'zrok://', 'file://')",
		Aliases: []string{"del"},
		Args:    cobra.ExactArgs(1),
	}
	command := &rmCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *rmCommand) run(_ *cobra.Command, args []string) {
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

	if err := target.Rm("/"); err != nil {
		tui.Error("error removing", err)
	}
}