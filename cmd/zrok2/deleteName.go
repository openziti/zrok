package main

import (
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/environment"
	"github.com/openziti/zrok/v2/rest_client_zrok/share"
	"github.com/spf13/cobra"
)

func init() {
	deleteCmd.AddCommand(newDeleteNameCommand().cmd)
}

type deleteNameCommand struct {
	cmd            *cobra.Command
	namespaceToken string
}

func newDeleteNameCommand() *deleteNameCommand {
	cmd := &cobra.Command{
		Use:   "name <name>",
		Short: "delete a name within a namespace",
		Args:  cobra.ExactArgs(1),
	}
	command := &deleteNameCommand{cmd: cmd}
	defaultNamespace := "public"
	if root, err := environment.LoadRoot(); err == nil {
		defaultNamespace, _ = root.DefaultNamespace()
	}
	cmd.Flags().StringVarP(&command.namespaceToken, "namespace-token", "n", defaultNamespace, "namespace token")
	cmd.Run = command.run
	return command
}

func (cmd *deleteNameCommand) run(_ *cobra.Command, args []string) {
	env, auth := mustGetEnvironmentAuth()

	zrok, err := env.Client()
	if err != nil {
		panic(err)
	}

	req := share.NewDeleteShareNameParams()
	req.Body = share.DeleteShareNameBody{
		NamespaceToken: cmd.namespaceToken,
		Name:           args[0],
	}

	_, err = zrok.Share.DeleteShareName(req, auth)
	if err != nil {
		panic(err)
	}

	dl.Infof("deleted name '%v' from namespace '%v'", args[0], cmd.namespaceToken)
}
