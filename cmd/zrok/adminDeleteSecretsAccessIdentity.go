package main

import (
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/rest_client_zrok"
	"github.com/spf13/cobra"
)

type adminDeleteSecretsAccessIdentityCommand struct {
	cmd *cobra.Command
}

func newDeleteSecretsAccessIdentityCommand() *adminDeleteSecretsAccessIdentityCommand {
	cmd := &cobra.Command{
		Use:     "secrets-access-identity <secretsAccessIdentityZId>",
		Aliases: []string{"sai"},
		Short:   "Delete a secrets access identity",
		Args:    cobra.ExactArgs(1),
	}
	command := &adminDeleteSecretsAccessIdentityCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *adminDeleteSecretsAccessIdentityCommand) run(_ *cobra.Command, args []string) {
	secretsAccessIdentityZId := args[0]

	env, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	zrok, err := env.Client()
	if err != nil {
		panic(err)
	}

	if err := cmd.deleteDialPolicy(secretsAccessIdentityZId, zrok); err != nil {
		panic(err)
	}

	if err := cmd.deleteIdentity(secretsAccessIdentityZId, zrok); err != nil {
		panic(err)
	}
}

func (cmd *adminDeleteSecretsAccessIdentityCommand) deleteDialPolicy(zId string, zrok *rest_client_zrok.Zrok) error {
	return nil
}

func (cmd *adminDeleteSecretsAccessIdentityCommand) deleteIdentity(zId string, zrok *rest_client_zrok.Zrok) error {
	return nil
}
