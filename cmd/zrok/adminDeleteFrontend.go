package main

import (
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok/admin"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/zrokdir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	adminDeleteCmd.AddCommand(newAdminDeleteFrontendCommand().cmd)
}

type adminDeleteFrontendCommand struct {
	cmd *cobra.Command
}

func newAdminDeleteFrontendCommand() *adminDeleteFrontendCommand {
	cmd := &cobra.Command{
		Use:     "frontend <frontendToken>",
		Aliases: []string{"fe"},
		Short:   "Delete a global public frontend",
		Args:    cobra.ExactArgs(1),
	}
	command := &adminDeleteFrontendCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *adminDeleteFrontendCommand) run(_ *cobra.Command, args []string) {
	feToken := args[0]

	zrok, err := zrokdir.ZrokClient(apiEndpoint)
	if err != nil {
		panic(err)
	}

	req := admin.NewDeleteFrontendParams()
	req.Body = &rest_model_zrok.DeleteFrontendRequest{FrontendToken: feToken}

	_, err = zrok.Admin.DeleteFrontend(req, mustGetAdminAuth())
	if err != nil {
		panic(err)
	}

	logrus.Infof("deleted global frontend '%v'", feToken)
}
