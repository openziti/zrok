package main

import (
	"os"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/rest_client_zrok/admin"
	"github.com/spf13/cobra"
)

func init() {
	adminCreateCmd.AddCommand(newAdminCreateNamespaceGrantCommand().cmd)
}

type adminCreateNamespaceGrantCommand struct {
	cmd *cobra.Command
}

func newAdminCreateNamespaceGrantCommand() *adminCreateNamespaceGrantCommand {
	cmd := &cobra.Command{
		Use:     "namespace-grant <namespaceToken> <accountEmail>",
		Aliases: []string{"ng"},
		Short:   "Grant an account access to a namespace",
		Args:    cobra.ExactArgs(2),
	}
	command := &adminCreateNamespaceGrantCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *adminCreateNamespaceGrantCommand) run(_ *cobra.Command, args []string) {
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

	req := admin.NewAddNamespaceGrantParams()
	req.Body.NamespaceToken = namespaceToken
	req.Body.Email = accountEmail

	if _, err = zrok.Admin.AddNamespaceGrant(req, mustGetAdminAuth()); err != nil {
		dl.Errorf("error adding namespace grant: %v", err)
		os.Exit(1)
	}

	dl.Infof("added namespace ('%v') grant for '%v'", namespaceToken, accountEmail)
}
