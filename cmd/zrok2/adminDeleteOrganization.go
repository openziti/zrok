package main

import (
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/environment"
	"github.com/openziti/zrok/v2/rest_client_zrok/admin"
	"github.com/spf13/cobra"
)

func init() {
	adminDeleteCmd.AddCommand(newAdminDeleteOrganizationCommand().cmd)
}

type adminDeleteOrganizationCommand struct {
	cmd *cobra.Command
}

func newAdminDeleteOrganizationCommand() *adminDeleteOrganizationCommand {
	cmd := &cobra.Command{
		Use:     "organization <organizationToken>",
		Aliases: []string{"org"},
		Short:   "Delete an organization",
		Args:    cobra.ExactArgs(1),
	}
	command := &adminDeleteOrganizationCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *adminDeleteOrganizationCommand) run(_ *cobra.Command, args []string) {
	env, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	zrok, err := env.Client()
	if err != nil {
		panic(err)
	}

	req := admin.NewDeleteOrganizationParams()
	req.Body.OrganizationToken = args[0]

	_, err = zrok.Admin.DeleteOrganization(req, mustGetAdminAuth())
	if err != nil {
		panic(err)
	}

	dl.Infof("deleted organization with token '%v'", args[0])
}
