package main

import (
	"os"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/rest_client_zrok/admin"
	"github.com/spf13/cobra"
)

func init() {
	adminDeleteCmd.AddCommand(newAdminDeleteFrontendGrantCommand().cmd)
}

type adminDeleteFrontendGrantCommand struct {
	cmd *cobra.Command
}

func newAdminDeleteFrontendGrantCommand() *adminDeleteFrontendGrantCommand {
	cmd := &cobra.Command{
		Use:     "frontend-grant <frontendToken> <accountEmail>",
		Aliases: []string{"fg"},
		Short:   "Remove account access from a frontend",
		Args:    cobra.ExactArgs(2),
	}
	command := &adminDeleteFrontendGrantCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *adminDeleteFrontendGrantCommand) run(_ *cobra.Command, args []string) {
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

	req := admin.NewDeleteFrontendGrantParams()
	req.Body.FrontendToken = frontendToken
	req.Body.Email = accountEmail

	if _, err := zrok.Admin.DeleteFrontendGrant(req, mustGetAdminAuth()); err != nil {
		dl.Errorf("error deleting frontend grant: %v", err)
		os.Exit(1)
	}

	dl.Infof("deleted frontend ('%v') grant for '%v'", frontendToken, accountEmail)
}
