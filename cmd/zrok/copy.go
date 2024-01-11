package main

import (
	"fmt"
	"github.com/openziti/zrok/drives/sync"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/openziti/zrok/tui"
	"github.com/spf13/cobra"
	"net/url"
	"os"
)

func init() {
	rootCmd.AddCommand(newCopyCommand().cmd)
}

type copyCommand struct {
	cmd       *cobra.Command
	sync      bool
	basicAuth string
}

func newCopyCommand() *copyCommand {
	cmd := &cobra.Command{
		Use:     "copy <source> [<target>] (<target> defaults to 'file://.`)",
		Short:   "Copy (unidirectional sync) zrok drive contents from <source> to <target> ('http://', 'file://', and 'zrok://' supported)",
		Aliases: []string{"cp"},
		Args:    cobra.RangeArgs(1, 2),
	}
	command := &copyCommand{cmd: cmd}
	cmd.Run = command.run
	cmd.Flags().BoolVarP(&command.sync, "sync", "s", false, "Only copy modified files (one-way synchronize)")
	cmd.Flags().StringVarP(&command.basicAuth, "basic-auth", "a", "", "Basic authentication <username:password>")
	return command
}

func (cmd *copyCommand) run(_ *cobra.Command, args []string) {
	if cmd.basicAuth == "" {
		cmd.basicAuth = os.Getenv("ZROK_DRIVES_BASIC_AUTH")
	}

	sourceUrl, err := url.Parse(args[0])
	if err != nil {
		tui.Error(fmt.Sprintf("invalid source '%v'", args[0]), err)
	}
	if sourceUrl.Scheme == "" {
		sourceUrl.Scheme = "file"
	}

	targetStr := "file://."
	if len(args) == 2 {
		targetStr = args[1]
	}
	targetUrl, err := url.Parse(targetStr)
	if err != nil {
		tui.Error(fmt.Sprintf("invalid target '%v'", targetStr), err)
	}
	if targetUrl.Scheme == "" {
		targetUrl.Scheme = "file"
	}

	root, err := environment.LoadRoot()
	if err != nil {
		tui.Error("error loading root", err)
	}

	var allocatedAccesses []*sdk.Access
	if sourceUrl.Scheme == "zrok" {
		access, err := sdk.CreateAccess(root, &sdk.AccessRequest{ShareToken: sourceUrl.Host})
		if err != nil {
			tui.Error("error creating access", err)
		}
		allocatedAccesses = append(allocatedAccesses, access)
	}
	if targetUrl.Scheme == "zrok" {
		access, err := sdk.CreateAccess(root, &sdk.AccessRequest{ShareToken: targetUrl.Host})
		if err != nil {
			tui.Error("error creating access", err)
		}
		allocatedAccesses = append(allocatedAccesses, access)
	}
	defer func() {
		for _, access := range allocatedAccesses {
			err := sdk.DeleteAccess(root, access)
			if err != nil {
				tui.Warning("error deleting target access", err)
			}
		}
	}()

	source, err := sync.TargetForURL(sourceUrl, root, cmd.basicAuth)
	if err != nil {
		tui.Error(fmt.Sprintf("error creating target for '%v'", sourceUrl), err)
	}
	target, err := sync.TargetForURL(targetUrl, root, cmd.basicAuth)
	if err != nil {
		tui.Error(fmt.Sprintf("error creating target for '%v'", targetUrl), err)
	}

	if err := sync.OneWay(source, target, cmd.sync); err != nil {
		tui.Error("error copying", err)
	}

	fmt.Println("copy complete!")
}
