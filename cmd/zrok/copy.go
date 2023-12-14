package main

import (
	"fmt"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/environment/env_core"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/openziti/zrok/tui"
	"github.com/openziti/zrok/util/sync"
	"github.com/pkg/errors"
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
	if sourceUrl.Scheme != "zrok" && sourceUrl.Scheme != "file" {
		tui.Error("source URL must be 'file://' or 'zrok://", nil)
	}

	targetStr := "file://."
	if len(args) == 2 {
		targetStr = args[1]
	}
	targetUrl, err := url.Parse(targetStr)
	if err != nil {
		tui.Error(fmt.Sprintf("invalid target URL '%v'", targetStr), err)
	}
	if targetUrl.Scheme != "zrok" && targetUrl.Scheme != "file" {
		tui.Error("target URL must be 'file://' or 'zrok://", nil)
	}

	if sourceUrl.Scheme != "zrok" && targetUrl.Scheme != "zrok" {
		tui.Error("either <source> or <target> must be a 'zrok://' URL", nil)
	}
	if targetUrl.Scheme != "file" && sourceUrl.Scheme != "file" {
		tui.Error("either <source> or <target> must be a 'file://' URL", nil)
	}

	root, err := environment.LoadRoot()
	if err != nil {
		tui.Error("error loading root", err)
	}

	var access *sdk.Access
	if sourceUrl.Scheme == "zrok" {
		access, err = sdk.CreateAccess(root, &sdk.AccessRequest{ShareToken: sourceUrl.Host})
		if err != nil {
			tui.Error("error creating access", err)
		}
	}
	if targetUrl.Scheme == "zrok" {
		access, err = sdk.CreateAccess(root, &sdk.AccessRequest{ShareToken: targetUrl.Host})
		if err != nil {
			tui.Error("error creating access", err)
		}
	}
	defer func() {
		err := sdk.DeleteAccess(root, access)
		if err != nil {
			tui.Error("error deleting access", err)
		}
	}()

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

func (cmd *copyCommand) createTarget(t *url.URL, root env_core.Root) (sync.Target, error) {
	switch t.Scheme {
	case "zrok":
		target, err := sync.NewWebDAVTarget(&sync.WebDAVTargetConfig{URL: t, Username: "", Password: "", Root: root})
		if err != nil {
			return nil, err
		}
		return target, nil

	case "file":
		return sync.NewFilesystemTarget(&sync.FilesystemTargetConfig{Root: t.Path}), nil

	default:
		return nil, errors.Errorf("invalid scheme")
	}
}
