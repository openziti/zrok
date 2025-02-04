package main

import (
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/rest_client_zrok/admin"
	"github.com/sirupsen/logrus"
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

	logrus.Infof("deleted organization with token '%v'", args[0])
}
