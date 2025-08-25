package main

import (
	"os"

	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/rest_client_zrok/admin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	adminDeleteCmd.AddCommand(newAdminDeleteNamespaceGrantCommand().cmd)
}

type adminDeleteNamespaceGrantCommand struct {
	cmd *cobra.Command
}

func newAdminDeleteNamespaceGrantCommand() *adminDeleteNamespaceGrantCommand {
	cmd := &cobra.Command{
		Use:     "namespace-grant <namespaceToken> <accountEmail>",
		Aliases: []string{"ng"},
		Short:   "Remove account access from a namespace",
		Args:    cobra.ExactArgs(2),
	}
	command := &adminDeleteNamespaceGrantCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *adminDeleteNamespaceGrantCommand) run(_ *cobra.Command, args []string) {
	namespaceToken := args[0]
	accountEmail := args[1]

	root, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	zrok, err := root.Client()
	if err != nil {
		panic(err)
	}

	req := admin.NewRemoveNamespaceGrantParams()
	req.Body.NamespaceToken = namespaceToken
	req.Body.Email = accountEmail

	if _, err := zrok.Admin.RemoveNamespaceGrant(req, mustGetAdminAuth()); err != nil {
		logrus.Errorf("error removing namespace grant: %v", err)
		os.Exit(1)
	}

	logrus.Infof("removed namespace ('%v') grant for '%v'", namespaceToken, accountEmail)
}