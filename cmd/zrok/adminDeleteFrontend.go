package main

import (
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/rest_client_zrok/admin"
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
		Use:   "frontend <frontendToken>",
		Short: "Delete a global public frontend",
		Args:  cobra.ExactArgs(1),
	}
	command := &adminDeleteFrontendCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *adminDeleteFrontendCommand) run(_ *cobra.Command, args []string) {
	feToken := args[0]

	env, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	zrok, err := env.Client()
	if err != nil {
		panic(err)
	}

	req := admin.NewDeleteFrontendParams()
	req.Body.FrontendToken = feToken

	if _, err := zrok.Admin.DeleteFrontend(req, mustGetAdminAuth()); err != nil {
		panic(err)
	}

	dl.Infof("deleted global frontend '%v'", feToken)
}
