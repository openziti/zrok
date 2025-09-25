package main

import (
	"os"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/rest_client_zrok/admin"
	"github.com/spf13/cobra"
)

func init() {
	adminCreateCmd.AddCommand(newAdminCreateNamespaceFrontendCommand().cmd)
}

type adminCreateNamespaceFrontendCommand struct {
	cmd       *cobra.Command
	isDefault bool
}

func newAdminCreateNamespaceFrontendCommand() *adminCreateNamespaceFrontendCommand {
	cmd := &cobra.Command{
		Use:     "namespace-frontend <namespaceToken> <frontendToken>",
		Aliases: []string{"fn"},
		Short:   "Map a frontend to a namespace",
		Args:    cobra.ExactArgs(2),
	}
	command := &adminCreateNamespaceFrontendCommand{cmd: cmd}
	cmd.Flags().BoolVar(&command.isDefault, "default", false, "create mapping as default")
	cmd.Run = command.run
	return command
}

func (cmd *adminCreateNamespaceFrontendCommand) run(_ *cobra.Command, args []string) {
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

	req := admin.NewAddNamespaceFrontendMappingParams()
	req.Body.FrontendToken = frontendToken
	req.Body.NamespaceToken = namespaceToken
	req.Body.IsDefault = cmd.isDefault

	if _, err = zrok.Admin.AddNamespaceFrontendMapping(req, mustGetAdminAuth()); err != nil {
		dl.Errorf("error creating namespace-frontend mapping: %v", err)
		os.Exit(1)
	}

	dl.Infof("created namespace-frontend mapping: namespace '%v' -> frontend '%v'", namespaceToken, frontendToken)
}
