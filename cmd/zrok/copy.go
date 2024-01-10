package main

import (
	"fmt"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/environment/env_core"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/openziti/zrok/tui"
	"github.com/openziti/zrok/util/sync"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/url"
)

func init() {
	rootCmd.AddCommand(newCopyCommand().cmd)
}

type copyCommand struct {
	cmd *cobra.Command
}

func newCopyCommand() *copyCommand {
	cmd := &cobra.Command{
		Use:   "copy <source> [<target>]",
		Short: "Copy zrok drive contents from <source> to <target> ('file://' and 'zrok://' supported)",
		Args:  cobra.RangeArgs(1, 2),
	}
	command := &copyCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *copyCommand) run(_ *cobra.Command, args []string) {
	sourceUrl, err := url.Parse(args[0])
	if err != nil {
		tui.Error(fmt.Sprintf("invalid source URL '%v'", args[0]), err)
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
		tui.Error(fmt.Sprintf("invalid target URL '%v'", targetStr), err)
	}
	if targetUrl.Scheme == "" {
		targetUrl.Scheme = "file"
	}

	root, err := environment.LoadRoot()
	if err != nil {
		tui.Error("error loading root", err)
	}

	var srcAccess *sdk.Access
	if sourceUrl.Scheme == "zrok" {
		srcAccess, err = sdk.CreateAccess(root, &sdk.AccessRequest{ShareToken: sourceUrl.Host})
		if err != nil {
			tui.Error("error creating access", err)
		}
	}
	if srcAccess != nil {
		defer func() {
			err := sdk.DeleteAccess(root, srcAccess)
			if err != nil {
				tui.Error("error deleting source access", err)
			}
		}()
	}
	var dstAccess *sdk.Access
	if targetUrl.Scheme == "zrok" {
		dstAccess, err = sdk.CreateAccess(root, &sdk.AccessRequest{ShareToken: targetUrl.Host})
		if err != nil {
			tui.Error("error creating access", err)
		}
	}
	if dstAccess != nil {
		defer func() {
			err := sdk.DeleteAccess(root, dstAccess)
			if err != nil {
				tui.Error("error deleting target access", err)
			}
		}()
	}

	source, err := cmd.createTarget(sourceUrl, root)
	if err != nil {
		tui.Error("error creating target", err)
	}
	target, err := cmd.createTarget(targetUrl, root)
	if err != nil {
		tui.Error("error creating target", err)
	}

	if err := sync.Synchronize(source, target); err != nil {
		tui.Error("error copying", err)
	}

	fmt.Println("copy complete!")
}

func (cmd *copyCommand) createTarget(url *url.URL, root env_core.Root) (sync.Target, error) {
	switch url.Scheme {
	case "file":
		logrus.Infof("%v", url)
		return sync.NewFilesystemTarget(&sync.FilesystemTargetConfig{Root: url.Path}), nil

	case "zrok":
		return sync.NewZrokTarget(&sync.ZrokTargetConfig{URL: url, Root: root})

	default:
		return sync.NewWebDAVTarget(&sync.WebDAVTargetConfig{URL: url, Username: "", Password: ""})
	}
}
