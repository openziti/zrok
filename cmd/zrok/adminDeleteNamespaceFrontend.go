package main

import (
	"os"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/environment"
	"github.com/openziti/zrok/v2/rest_client_zrok/admin"
	"github.com/spf13/cobra"
)

func init() {
	adminDeleteCmd.AddCommand(newAdminDeleteNamespaceFrontendCommand().cmd)
}

type adminDeleteNamespaceFrontendCommand struct {
	cmd *cobra.Command
}

func newAdminDeleteNamespaceFrontendCommand() *adminDeleteNamespaceFrontendCommand {
	cmd := &cobra.Command{
		Use:     "namespace-frontend <namespaceToken> <frontendToken>",
		Aliases: []string{"fn"},
		Short:   "Remove mapping between frontend and namespace",
		Args:    cobra.ExactArgs(2),
	}
	command := &adminDeleteNamespaceFrontendCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *adminDeleteNamespaceFrontendCommand) run(_ *cobra.Command, args []string) {
	namespaceToken := args[0]
	frontendToken := args[1]

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
		dl.Errorf("error deleting namespace-frontend mapping: %v", err)
		os.Exit(1)
	}

	dl.Infof("deleted namespace-frontend mapping: namespace '%v' -> frontend '%v'", namespaceToken, frontendToken)
}
