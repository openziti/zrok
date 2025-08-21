package main

import (
	"os"

	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/rest_client_zrok/admin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	adminCreateCmd.AddCommand(newAdminCreateFrontendNamespaceCommand().cmd)
}

type adminCreateFrontendNamespaceCommand struct {
	cmd       *cobra.Command
	isDefault bool
}

func newAdminCreateFrontendNamespaceCommand() *adminCreateFrontendNamespaceCommand {
	cmd := &cobra.Command{
		Use:     "frontend-namespace <frontendToken> <namespaceToken>",
		Aliases: []string{"fn"},
		Short:   "Map a frontend to a namespace",
		Args:    cobra.ExactArgs(2),
	}
	command := &adminCreateFrontendNamespaceCommand{cmd: cmd}
	cmd.Flags().BoolVar(&command.isDefault, "default", false, "create mapping as default")
	cmd.Run = command.run
	return command
}

func (cmd *adminCreateFrontendNamespaceCommand) run(_ *cobra.Command, args []string) {
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

	req := admin.NewAddNamespaceFrontendMappingParams()
	req.Body.FrontendToken = frontendToken
	req.Body.NamespaceToken = namespaceToken
	req.Body.IsDefault = cmd.isDefault

	if _, err = zrok.Admin.AddNamespaceFrontendMapping(req, mustGetAdminAuth()); err != nil {
		logrus.Errorf("error creating frontend-namespace mapping: %v", err)
		os.Exit(1)
	}

	logrus.Infof("created frontend-namespace mapping: frontend '%v' -> namespace '%v'", frontendToken, namespaceToken)
}