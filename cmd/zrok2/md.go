package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/drives/sync"
	"github.com/openziti/zrok/v2/environment"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
	"github.com/openziti/zrok/v2/tui"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newMdCommand().cmd)
}

type mdCommand struct {
	cmd       *cobra.Command
	basicAuth string
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
	cmd.Flags().StringVarP(&command.basicAuth, "basic-auth", "a", "", "Basic authentication <username:password>")
	return command
}

func (cmd *mdCommand) run(_ *cobra.Command, args []string) {
	if cmd.basicAuth == "" {
		cmd.basicAuth = os.Getenv("ZROK2_DRIVES_BASIC_AUTH")
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

	if err := target.Mkdir("/"); err != nil {
		tui.Error("error creating directory", err)
	}
}
