package main

import (
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/rest_client_zrok/admin"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/openziti/zrok/tui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	adminCreateCmd.AddCommand(newAdminCreateFrontendCommand().cmd)
}

type adminCreateFrontendCommand struct {
	cmd    *cobra.Command
	closed bool
}

func newAdminCreateFrontendCommand() *adminCreateFrontendCommand {
	cmd := &cobra.Command{
		Use:   "frontend <zitiId> <publicName> <urlTemplate>",
		Short: "Create a global public frontend",
		Args:  cobra.ExactArgs(3),
	}
	command := &adminCreateFrontendCommand{cmd: cmd}
	cmd.Flags().BoolVar(&command.closed, "closed", false, "Enabled closed permission mode")
	cmd.Run = command.run
	return command
}

func (cmd *adminCreateFrontendCommand) run(_ *cobra.Command, args []string) {
	zId := args[0]
	publicName := args[1]
	urlTemplate := args[2]

	env, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	zrok, err := env.Client()
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

	resp, err := zrok.Admin.CreateFrontend(req, mustGetAdminAuth())
	if err != nil {
		switch err.(type) {
		case *admin.CreateFrontendBadRequest:
			tui.Error("create frontend request failed: name already exists", err)
			os.Exit(1)
		default:
			tui.Error("create frontend request failed", err)
			os.Exit(1)
		}
	}

	logrus.Infof("created global public frontend '%v'", resp.Payload.FrontendToken)
}
