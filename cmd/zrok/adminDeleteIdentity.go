package main

import (
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/rest_client_zrok/admin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	adminDeleteCmd.AddCommand(newAdminDeleteIdentityCommand().cmd)
}

type adminDeleteIdentityCommand struct {
	cmd *cobra.Command
}

func newAdminDeleteIdentityCommand() *adminDeleteIdentityCommand {
	cmd := &cobra.Command{
		Use:   "identity <zId>",
		Short: "Delete an identity",
		Args:  cobra.ExactArgs(1),
	}
	command := &adminDeleteIdentityCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *adminDeleteIdentityCommand) run(_ *cobra.Command, args []string) {
	zId := args[0]

	env, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	zrok, err := env.Client()
	if err != nil {
		panic(err)
	}

	req := admin.NewDeleteIdentityParams()
	req.Body.ZID = zId

	if _, err := zrok.Admin.DeleteIdentity(req, mustGetAdminAuth()); err != nil {
		panic(err)
	}

	logrus.Infof("deleted identity '%v'; please remove any related identity json files", zId)
}
