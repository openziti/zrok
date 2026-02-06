package main

import (
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/environment"
	"github.com/openziti/zrok/v2/rest_client_zrok/admin"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
	"github.com/openziti/zrok/v2/tui"
	"github.com/spf13/cobra"
)

func init() {
	adminCreateCmd.AddCommand(newAdminCreateFrontendCommand().cmd)
}

type adminCreateFrontendCommand struct {
	cmd     *cobra.Command
	closed  bool
	dynamic bool
}

func newAdminCreateFrontendCommand() *adminCreateFrontendCommand {
	cmd := &cobra.Command{
		Use:   "frontend <zitiId> <publicName> <urlTemplate>",
		Short: "Create a global public frontend",
		Args:  cobra.ExactArgs(3),
	}
	command := &adminCreateFrontendCommand{cmd: cmd}
	cmd.Flags().BoolVar(&command.closed, "closed", false, "Enabled closed permission mode")
	cmd.Flags().BoolVar(&command.dynamic, "dynamic", false, "Enable dynamic mode for the frontend")
	cmd.Run = command.run
	return command
}

func (cmd *adminCreateFrontendCommand) run(_ *cobra.Command, args []string) {
	zId := args[0]
	publicName := args[1]
	urlTemplate := args[2]

	root, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	zrok, err := root.Client()
	if err != nil {
		panic(err)
	}

	permissionMode := sdk.OpenPermissionMode
	if cmd.closed {
		permissionMode = sdk.ClosedPermissionMode
	}
	req := admin.NewCreateFrontendParams()
	req.Body.ZID = zId
	req.Body.PublicName = publicName
	req.Body.URLTemplate = urlTemplate
	req.Body.PermissionMode = string(permissionMode)
	req.Body.Dynamic = cmd.dynamic

	resp, err := zrok.Admin.CreateFrontend(req, mustGetAdminAuth())
	if err != nil {
		switch err.(type) {
		case *admin.CreateFrontendBadRequest:
			tui.Error("create frontend request failed: name already exists", err)
		default:
			tui.Error("create frontend request failed", err)
		}
	}

	dl.Infof("created global public frontend '%v'", resp.Payload.FrontendToken)
}
