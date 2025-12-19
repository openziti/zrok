package main

import (
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/environment"
	"github.com/openziti/zrok/v2/rest_client_zrok/admin"
	"github.com/spf13/cobra"
)

func init() {
	adminDeleteCmd.AddCommand(newAdminDeleteAccountCommand().cmd)
}

type adminDeleteAccountCommand struct {
	cmd *cobra.Command
}

func newAdminDeleteAccountCommand() *adminDeleteAccountCommand {
	cmd := &cobra.Command{
		Use:   "account <email>",
		Short: "Delete an account and disable all allocated resources",
		Args:  cobra.ExactArgs(1),
	}
	command := &adminDeleteAccountCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *adminDeleteAccountCommand) run(_ *cobra.Command, args []string) {
	email := args[0]

	root, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	zrok, err := root.Client()
	if err != nil {
		panic(err)
	}

	req := admin.NewDeleteAccountParams()
	req.Body.Email = email

	if _, err := zrok.Admin.DeleteAccount(req, mustGetAdminAuth()); err != nil {
		panic(err)
	}

	dl.Infof("deleted account '%v'", email)
}
