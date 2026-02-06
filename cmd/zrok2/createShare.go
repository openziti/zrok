package main

import (
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/environment"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
	"github.com/openziti/zrok/v2/tui"
	"github.com/spf13/cobra"
)

func init() {
	createCmd.AddCommand(newCreateShareCommand().cmd)
}

type createShareCommand struct {
	cmd          *cobra.Command
	backendMode  string
	shareToken   string
	open         bool
	accessGrants []string
}

func newCreateShareCommand() *createShareCommand {
	cmd := &cobra.Command{
		Use:   "share",
		Short: "Create a private share without starting a backend",
		Args:  cobra.NoArgs,
	}
	command := &createShareCommand{cmd: cmd}
	cmd.Flags().StringVarP(&command.backendMode, "backend-mode", "b", "proxy", "The backend mode {proxy, web, tcpTunnel, udpTunnel, caddy, drive, socks}")
	cmd.Flags().StringVarP(&command.shareToken, "share-token", "s", "", "Request a specific share token name")
	cmd.Flags().BoolVar(&command.open, "open", false, "Enable open permission mode")
	cmd.Flags().StringArrayVar(&command.accessGrants, "access-grant", []string{}, "zrok accounts that are allowed to access this share")
	cmd.Run = command.run
	return command
}

func (cmd *createShareCommand) run(_ *cobra.Command, _ []string) {
	root, err := environment.LoadRoot()
	if err != nil {
		tui.Error("error loading environment", err)
	}

	if !root.IsEnabled() {
		tui.Error("unable to load environment; did you 'zrok2 enable'?", nil)
	}

	req := &sdk.ShareRequest{
		BackendMode:       sdk.BackendMode(cmd.backendMode),
		ShareMode:         sdk.PrivateShareMode,
		PrivateShareToken: cmd.shareToken,
		PermissionMode:    sdk.ClosedPermissionMode,
		AccessGrants:      cmd.accessGrants,
	}
	if cmd.open {
		req.PermissionMode = sdk.OpenPermissionMode
	}

	shr, err := sdk.CreateShare(root, req)
	if err != nil {
		tui.Error("unable to create share", err)
	}

	dl.Infof("created share '%v'", shr.Token)
}
