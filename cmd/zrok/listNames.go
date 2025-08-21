package main

import (
	"fmt"

	"github.com/openziti/zrok/rest_client_zrok/share"
	"github.com/spf13/cobra"
)

func init() {
	listCmd.AddCommand(newListNamesCommand().cmd)
}

type listNamesCommand struct {
	cmd            *cobra.Command
	namespaceToken string
}

func newListNamesCommand() *listNamesCommand {
	cmd := &cobra.Command{
		Use:   "names",
		Short: "list names within a namespace",
		Args:  cobra.NoArgs,
	}
	command := &listNamesCommand{cmd: cmd}
	cmd.Flags().StringVarP(&command.namespaceToken, "namespace-token", "n", "", "namespace token")
	cmd.MarkFlagRequired("namespace-token")
	cmd.Run = command.run
	return command
}

func (cmd *listNamesCommand) run(_ *cobra.Command, args []string) {
	env, auth := mustGetEnvironmentAuth()

	zrok, err := env.Client()
	if err != nil {
		panic(err)
	}

	req := share.NewListShareNamesParams()
	req.NamespaceToken = cmd.namespaceToken

	resp, err := zrok.Share.ListShareNames(req, auth)
	if err != nil {
		panic(err)
	}

	for _, name := range resp.Payload {
		fmt.Println(name.Name)
	}
}
