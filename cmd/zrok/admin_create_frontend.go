package main

import (
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok/admin"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/zrokdir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	adminCreateCmd.AddCommand(newAdminCreateFrontendCommand().cmd)
}

type adminCreateFrontendCommand struct {
	cmd *cobra.Command
}

func newAdminCreateFrontendCommand() *adminCreateFrontendCommand {
	cmd := &cobra.Command{
		Use:   "frontend <zitiId> <publicName> <urlTemplate>",
		Short: "Create a global public frontend",
		Args:  cobra.ExactArgs(3),
	}
	command := &adminCreateFrontendCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *adminCreateFrontendCommand) run(_ *cobra.Command, args []string) {
	zId := args[0]
	publicName := args[1]
	urlTemplate := args[2]

	zrok, err := zrokdir.ZrokClient(apiEndpoint)
	if err != nil {
		panic(err)
	}

	req := admin.NewCreateFrontendParams()
	req.Body = &rest_model_zrok.CreateFrontendRequest{
		ZID:         zId,
		PublicName:  publicName,
		URLTemplate: urlTemplate,
	}

	resp, err := zrok.Admin.CreateFrontend(req, mustGetAdminAuth())
	if err != nil {
		panic(err)
	}

	logrus.Infof("created global public frontend '%v'", resp.Payload.Token)
}
