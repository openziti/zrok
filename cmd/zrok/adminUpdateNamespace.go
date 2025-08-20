package main

import (
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/rest_client_zrok/admin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	adminUpdateCmd.AddCommand(newAdminUpdateNamespaceCommand().cmd)
}

type adminUpdateNamespaceCommand struct {
	cmd         *cobra.Command
	name        string
	description string
}

func newAdminUpdateNamespaceCommand() *adminUpdateNamespaceCommand {
	cmd := &cobra.Command{
		Use:   "namespace <token>",
		Short: "Update a namespace",
		Args:  cobra.ExactArgs(1),
	}
	command := &adminUpdateNamespaceCommand{cmd: cmd}
	cmd.Flags().StringVarP(&command.name, "name", "n", "", "namespace name")
	cmd.Flags().StringVarP(&command.description, "description", "d", "", "namespace description")
	cmd.Run = command.run
	return command
}

func (cmd *adminUpdateNamespaceCommand) run(_ *cobra.Command, args []string) {
	env, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	zrok, err := env.Client()
	if err != nil {
		panic(err)
	}

	req := admin.NewUpdateNamespaceParams()
	req.Body = admin.UpdateNamespaceBody{
		NamespaceToken: args[0],
		Name:           cmd.name,
		Description:    cmd.description,
	}

	_, err = zrok.Admin.UpdateNamespace(req, mustGetAdminAuth())
	if err != nil {
		panic(err)
	}

	logrus.Infof("updated namespace '%v'", args[0])
}