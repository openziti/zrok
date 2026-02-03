package main

import (
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/environment"
	"github.com/openziti/zrok/v2/rest_client_zrok/share"
	"github.com/openziti/zrok/v2/tui"
	"github.com/spf13/cobra"
)

func init() {
	createCmd.AddCommand(newCreateNameCommand().cmd)
}

type createNameCommand struct {
	cmd            *cobra.Command
	namespaceToken string
}

func newCreateNameCommand() *createNameCommand {
	cmd := &cobra.Command{
		Use:   "name <name>",
		Short: "create a name within a namespace",
		Args:  cobra.ExactArgs(1),
	}
	command := &createNameCommand{cmd: cmd}
	defaultNamespace := "public"
	if root, err := environment.LoadRoot(); err == nil {
		defaultNamespace, _ = root.DefaultNamespace()
	}
	cmd.Flags().StringVarP(&command.namespaceToken, "namespace-token", "n", defaultNamespace, "namespace token")
	cmd.Run = command.run
	return command
}

func (cmd *createNameCommand) run(_ *cobra.Command, args []string) {
	env, auth := mustGetEnvironmentAuth()

	zrok, err := env.Client()
	if err != nil {
		tui.Error("unable to get zrok client", err)
	}

	req := share.NewCreateShareNameParams()
	req.Body = share.CreateShareNameBody{
		NamespaceToken: cmd.namespaceToken,
		Name:           args[0],
	}

	_, err = zrok.Share.CreateShareName(req, auth)
	if err != nil {
		tui.Error("unable to create name", err)
	}

	dl.Infof("created name '%v' in namespace '%v'", args[0], cmd.namespaceToken)
}
