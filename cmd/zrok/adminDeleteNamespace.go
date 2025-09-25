package main

import (
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/rest_client_zrok/admin"
	"github.com/spf13/cobra"
)

func init() {
	adminDeleteCmd.AddCommand(newAdminDeleteNamespaceCommand().cmd)
}

type adminDeleteNamespaceCommand struct {
	cmd *cobra.Command
}

func newAdminDeleteNamespaceCommand() *adminDeleteNamespaceCommand {
	cmd := &cobra.Command{
		Use:   "namespace <token>",
		Short: "Delete a namespace",
		Args:  cobra.ExactArgs(1),
	}
	command := &adminDeleteNamespaceCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *adminDeleteNamespaceCommand) run(_ *cobra.Command, args []string) {
	env, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	zrok, err := env.Client()
	if err != nil {
		panic(err)
	}

	req := admin.NewDeleteNamespaceParams()
	req.Body = admin.DeleteNamespaceBody{
		NamespaceToken: args[0],
	}

	_, err = zrok.Admin.DeleteNamespace(req, mustGetAdminAuth())
	if err != nil {
		panic(err)
	}

	dl.Infof("deleted namespace '%v'", args[0])
}
