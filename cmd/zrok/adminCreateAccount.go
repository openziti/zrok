package main

import (
	"fmt"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/rest_client_zrok/admin"
	"github.com/spf13/cobra"
)

func init() {
	adminCreateCmd.AddCommand(newAdminCreateAccount().cmd)
}

type adminCreateAccount struct {
	cmd *cobra.Command
}

func newAdminCreateAccount() *adminCreateAccount {
	cmd := &cobra.Command{
		Use:   "account <email> <password>",
		Short: "Pre-populate an account in the database; returns an enable token for the account",
		Args:  cobra.ExactArgs(2),
	}
	command := &adminCreateAccount{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *adminCreateAccount) run(_ *cobra.Command, args []string) {
	env, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	zrok, err := env.Client()
	if err != nil {
		panic(err)
	}

	req := admin.NewCreateAccountParams()
	req.Body.Email = args[0]
	req.Body.Password = args[1]

	resp, err := zrok.Admin.CreateAccount(req, mustGetAdminAuth())
	if err != nil {
		panic(err)
	}

	fmt.Println(resp.GetPayload().AccountToken)
}
