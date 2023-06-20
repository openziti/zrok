package main

import (
	"os"

	"github.com/openziti/zrok/rest_client_zrok/admin"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/tui"
	"github.com/openziti/zrok/zrokdir"
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

	zrd, err := zrokdir.Load()
	if err != nil {
		panic(err)
	}

	zrok, err := zrd.Client()
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
		switch err.(type) {
		case *admin.CreateFrontendBadRequest:
			tui.Error("create frontend request failed: name already exists", err)
			os.Exit(1)
		default:
			tui.Error("create frontend request failed", err)
			os.Exit(1)
		}
	}

	logrus.Infof("created global public frontend '%v'", resp.Payload.Token)
}
