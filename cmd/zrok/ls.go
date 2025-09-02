package main

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/openziti/zrok/drives/sync"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/openziti/zrok/tui"
	"github.com/openziti/zrok/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/url"
	"os"
	"sort"
)

func init() {
	rootCmd.AddCommand(newLsCommand().cmd)
}

type lsCommand struct {
	cmd       *cobra.Command
	basicAuth string
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
	cmd.Flags().StringVarP(&command.basicAuth, "basic-auth", "a", "", "Basic authentication <username:password>")
	return command
}

func (cmd *lsCommand) run(_ *cobra.Command, args []string) {
	if cmd.basicAuth == "" {
		cmd.basicAuth = os.Getenv("ZROK_DRIVES_BASIC_AUTH")
	}

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

	if targetUrl.Scheme == "zrok" {
		access, err := sdk.CreateAccess(root, &sdk.AccessRequest{ShareToken: targetUrl.Host})
		if err != nil {
			tui.Error("error creating access", err)
		}
		defer func() {
			if err := sdk.DeleteAccess(root, access); err != nil {
				logrus.Warningf("error freeing access: %v", err)
			}
		}()
	}

	target, err := sync.TargetForURL(targetUrl, root, cmd.basicAuth)
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
	tw.SetStyle(table.StyleRounded)
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
