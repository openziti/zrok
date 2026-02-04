package main

import (
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/environment"
	"github.com/openziti/zrok/v2/rest_client_zrok/admin"
	"github.com/spf13/cobra"
)

func init() {
	adminCreateCmd.AddCommand(newAdminCreateNamespaceCommand().cmd)
}

type adminCreateNamespaceCommand struct {
	cmd         *cobra.Command
	token       string
	description string
	open        bool
}

func newAdminCreateNamespaceCommand() *adminCreateNamespaceCommand {
	cmd := &cobra.Command{
		Use:   "namespace <name>",
		Short: "Create a new namespace",
		Args:  cobra.ExactArgs(1),
	}
	command := &adminCreateNamespaceCommand{cmd: cmd}
	cmd.Flags().StringVarP(&command.token, "token", "t", "", "provide a custom token")
	cmd.Flags().StringVarP(&command.description, "description", "d", "", "namespace description")
	cmd.Flags().BoolVarP(&command.open, "open", "o", false, "namespace is open to all accounts (default is closed)")
	cmd.Run = command.run
	return command
}

func (cmd *adminCreateNamespaceCommand) run(_ *cobra.Command, args []string) {
	env, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	zrok, err := env.Client()
	if err != nil {
		panic(err)
	}

	req := admin.NewCreateNamespaceParams()
	req.Body = admin.CreateNamespaceBody{
		Name:        args[0],
		Description: cmd.description,
		Open:        cmd.open,
	}
	if cmd.token != "" {
		req.Body.Token = cmd.token
	}

	resp, err := zrok.Admin.CreateNamespace(req, mustGetAdminAuth())
	if err != nil {
		panic(err)
	}

	dl.Infof("created namespace '%v' with token '%v'", args[0], resp.Payload.NamespaceToken)
}
