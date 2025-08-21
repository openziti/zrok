package main

import (
	"github.com/openziti/zrok/rest_client_zrok/share"
	"github.com/sirupsen/logrus"
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
	cmd.Flags().StringVarP(&command.namespaceToken, "namespace-token", "n", "", "namespace token")
	cmd.MarkFlagRequired("namespace-token")
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

	logrus.Infof("deleted name '%v' from namespace", args[0])
}
