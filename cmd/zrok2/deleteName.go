package main

import (
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/environment"
	"github.com/openziti/zrok/v2/rest_client_zrok/share"
	"github.com/openziti/zrok/v2/tui"
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
		tui.Error("unable to get zrok client", err)
	}

	req := share.NewDeleteShareNameParams()
	req.Body = share.DeleteShareNameBody{
		NamespaceToken: cmd.namespaceToken,
		Name:           args[0],
	}

	_, err = zrok.Share.DeleteShareName(req, auth)
	if err != nil {
		if conflict, ok := err.(*share.DeleteShareNameConflict); ok {
			tui.Error(string(conflict.GetPayload()), nil)
		}
		tui.Error("unable to delete name", err)
	}

	dl.Infof("deleted name '%v' from namespace '%v'", args[0], cmd.namespaceToken)
}
