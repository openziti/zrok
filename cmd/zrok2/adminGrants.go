package main

import (
	"fmt"
	"github.com/openziti/zrok/v2/environment"
	"github.com/openziti/zrok/v2/rest_client_zrok/admin"
	"github.com/spf13/cobra"
)

func init() {
	adminCmd.AddCommand(newAdminGrantsCommand().cmd)
}

type adminGrantsCommand struct {
	cmd   *cobra.Command
	email string
}

func newAdminGrantsCommand() *adminGrantsCommand {
	cmd := &cobra.Command{
		Use:   "grants <email>",
		Short: "Synchronize ziti objects with account grants",
		Args:  cobra.ExactArgs(1),
	}
	command := &adminGrantsCommand{cmd: cmd}
	cmd.Run = command.run
	cmd.Flags().StringVarP(&command.email, "email", "e", "", "email address")
	return command
}

func (command *adminGrantsCommand) run(_ *cobra.Command, args []string) {
	env, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	zrok, err := env.Client()
	if err != nil {
		panic(err)
	}

	req := admin.NewGrantsParams()
	req.Body.Email = args[0]

	_, err = zrok.Admin.Grants(req, mustGetAdminAuth())
	if err != nil {
		panic(err)
	}

	fmt.Println("success.")
}
