package main

import (
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/rest_client_zrok/admin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	adminCreateCmd.AddCommand(newAdminCreateNamespaceCommand().cmd)
}

type adminCreateNamespaceCommand struct {
	cmd         *cobra.Command
	name        string
	description string
}

func newAdminCreateNamespaceCommand() *adminCreateNamespaceCommand {
	cmd := &cobra.Command{
		Use:   "namespace <name>",
		Short: "Create a new namespace",
		Args:  cobra.ExactArgs(1),
	}
	command := &adminCreateNamespaceCommand{cmd: cmd}
	cmd.Flags().StringVarP(&command.description, "description", "d", "", "namespace description")
	cmd.Run = command.run
	return command
}

func (cmd *adminCreateNamespaceCommand) run(_ *cobra.Command, args []string) {
	env, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	zrok, err := env.Client()
	if err != nil {
		panic(err)
	}

	req := admin.NewCreateNamespaceParams()
	req.Body = admin.CreateNamespaceBody{
		Name:        args[0],
		Description: cmd.description,
	}

	resp, err := zrok.Admin.CreateNamespace(req, mustGetAdminAuth())
	if err != nil {
		panic(err)
	}

	logrus.Infof("created namespace '%v' with token '%v'", args[0], resp.Payload.NamespaceToken)
}