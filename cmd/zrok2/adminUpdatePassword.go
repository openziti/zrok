package main

import (
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/environment"
	"github.com/openziti/zrok/v2/rest_client_zrok/admin"
	"github.com/spf13/cobra"
)

func init() {
	adminUpdateCmd.AddCommand(newAdminUpdatePasswordCommand().cmd)
}

type adminUpdatePasswordCommand struct {
	cmd *cobra.Command
}

func newAdminUpdatePasswordCommand() *adminUpdatePasswordCommand {
	cmd := &cobra.Command{
		Use:   "password <email> <password>",
		Short: "Update the password of an account",
		Args:  cobra.ExactArgs(2),
	}
	command := &adminUpdatePasswordCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *adminUpdatePasswordCommand) run(_ *cobra.Command, args []string) {
	env, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	zrok, err := env.Client()
	if err != nil {
		panic(err)
	}

	req := admin.NewUpdateAccountPasswordParams()
	req.Body.Email = args[0]
	req.Body.Password = args[1]

	_, err = zrok.Admin.UpdateAccountPassword(req, mustGetAdminAuth())
	if err != nil {
		panic(err)
	}

	dl.Infof("updated password for '%v'", args[0])
}
