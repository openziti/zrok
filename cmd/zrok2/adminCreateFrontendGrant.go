package main

import (
	"os"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/environment"
	"github.com/openziti/zrok/v2/rest_client_zrok/admin"
	"github.com/spf13/cobra"
)

func init() {
	adminCreateCmd.AddCommand(newAdminCreateFrontendGrantCommand().cmd)
}

type adminCreateFrontendGrantCommand struct {
	cmd *cobra.Command
}

func newAdminCreateFrontendGrantCommand() *adminCreateFrontendGrantCommand {
	cmd := &cobra.Command{
		Use:     "frontend-grant <frontendToken> <accountEmail>",
		Aliases: []string{"fg"},
		Short:   "Grant an account access to a frontend",
		Args:    cobra.ExactArgs(2),
	}
	command := &adminCreateFrontendGrantCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *adminCreateFrontendGrantCommand) run(_ *cobra.Command, args []string) {
	frontendToken := args[0]
	accountEmail := args[1]

	root, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	zrok, err := root.Client()
	if err != nil {
		panic(err)
	}

	req := admin.NewAddFrontendGrantParams()
	req.Body.FrontendToken = frontendToken
	req.Body.Email = accountEmail

	if _, err = zrok.Admin.AddFrontendGrant(req, mustGetAdminAuth()); err != nil {
		dl.Errorf("error addming frontend grant: %v", err)
		os.Exit(1)
	}

	dl.Infof("added frontend ('%v') grant for '%v'", frontendToken, accountEmail)
}
