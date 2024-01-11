package main

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/openziti/zrok/drives/sync"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/tui"
	"github.com/openziti/zrok/util"
	"github.com/spf13/cobra"
	"net/url"
	"os"
	"sort"
)

func init() {
	rootCmd.AddCommand(newLsCommand().cmd)
}

type lsCommand struct {
	cmd *cobra.Command
}

func newLsCommand() *lsCommand {
	cmd := &cobra.Command{
		Use:     "ls <target>",
		Short:   "List the contents of drive <target> ('http://', 'zrok://','file://')",
		Aliases: []string{"dir"},
		Args:    cobra.ExactArgs(1),
	}
	command := &lsCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *lsCommand) run(_ *cobra.Command, args []string) {
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

	objects, err := target.Dir("/")
	if err != nil {
		tui.Error("error listing directory", err)
	}
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].Path < objects[j].Path
	})

	tw := table.NewWriter()
	tw.SetOutputMirror(os.Stdout)
	tw.SetStyle(table.StyleLight)
	tw.AppendHeader(table.Row{"type", "Name", "Size", "Modified"})
	for _, object := range objects {
		if object.IsDir {
			tw.AppendRow(table.Row{"DIR", object.Path, "", ""})
		} else {
			tw.AppendRow(table.Row{"", object.Path, util.BytesToSize(object.Size), object.Modified.Local()})
		}
	}
	tw.Render()
}
