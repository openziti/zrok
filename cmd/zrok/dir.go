package main

import (
	"fmt"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/tui"
	"github.com/openziti/zrok/util"
	"github.com/openziti/zrok/util/sync"
	"github.com/spf13/cobra"
	"net/url"
)

func init() {
	rootCmd.AddCommand(newDirCommand().cmd)
}

type dirCommand struct {
	cmd *cobra.Command
}

func newDirCommand() *dirCommand {
	cmd := &cobra.Command{
		Use:     "dir <target>",
		Short:   "List the contents of <target> ('http://', 'zrok://', and 'file://' supported)",
		Aliases: []string{"ls"},
		Args:    cobra.ExactArgs(1),
	}
	command := &dirCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *dirCommand) run(_ *cobra.Command, args []string) {
	targetUrl, err := url.Parse(args[0])
	if err != nil {
		tui.Error(fmt.Sprintf("invalid target URL '%v'", args[0]), err)
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
		tui.Error(fmt.Sprintf("error creating target for '%v'", targetUrl.String()), err)
	}

	objects, err := target.Dir("/")
	if err != nil {
		tui.Error("error listing directory", err)
	}

	for _, object := range objects {
		if object.IsDir {
			fmt.Printf("<dir> %-32s\n", object.Path)
		} else {
			fmt.Printf("      %-32s %-16s %s\n", object.Path, util.BytesToSize(object.Size), object.Modified.String())
		}
	}
	fmt.Println()
}
