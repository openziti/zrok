package main

import (
	"os"

	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/rest_client_zrok/admin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	adminDeleteCmd.AddCommand(newAdminDeleteFrontendNamespaceCommand().cmd)
}

type adminDeleteFrontendNamespaceCommand struct {
	cmd *cobra.Command
}

func newAdminDeleteFrontendNamespaceCommand() *adminDeleteFrontendNamespaceCommand {
	cmd := &cobra.Command{
		Use:     "frontend-namespace <frontendToken> <namespaceToken>",
		Aliases: []string{"fn"},
		Short:   "Remove mapping between frontend and namespace",
		Args:    cobra.ExactArgs(2),
	}
	command := &adminDeleteFrontendNamespaceCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *adminDeleteFrontendNamespaceCommand) run(_ *cobra.Command, args []string) {
	frontendToken := args[0]
	namespaceToken := args[1]

	root, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	zrok, err := root.Client()
	if err != nil {
		panic(err)
	}

	req := admin.NewRemoveNamespaceFrontendMappingParams()
	req.Body.FrontendToken = frontendToken
	req.Body.NamespaceToken = namespaceToken

	if _, err := zrok.Admin.RemoveNamespaceFrontendMapping(req, mustGetAdminAuth()); err != nil {
		logrus.Errorf("error deleting frontend-namespace mapping: %v", err)
		os.Exit(1)
	}

	logrus.Infof("deleted frontend-namespace mapping: frontend '%v' -> namespace '%v'", frontendToken, namespaceToken)
}