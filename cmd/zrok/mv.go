package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/drives/sync"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/openziti/zrok/tui"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newMvCommand().cmd)
}

type mvCommand struct {
	cmd       *cobra.Command
	basicAuth string
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
	cmd.Flags().StringVarP(&command.basicAuth, "basic-auth", "a", "", "Basic authentication <username:password>")
	return command
}

func (cmd *mvCommand) run(_ *cobra.Command, args []string) {
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
				dl.Warnf("error freeing access: %v", err)
			}
		}()
	}

	target, err := sync.TargetForURL(targetUrl, root, cmd.basicAuth)
	if err != nil {
		tui.Error(fmt.Sprintf("error creating target for '%v'", targetUrl), err)
	}

	if err := target.Move("/", args[1]); err != nil {
		tui.Error("error moving", err)
	}
}
