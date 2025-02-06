package main

import (
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/rest_client_zrok/admin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	adminDeleteCmd.AddCommand(newAdminDeleteOrgMemberCommand().cmd)
}

type adminDeleteOrgMemberCommand struct {
	cmd *cobra.Command
}

func newAdminDeleteOrgMemberCommand() *adminDeleteOrgMemberCommand {
	cmd := &cobra.Command{
		Use:     "org-member <organizationToken> <accountEmail>",
		Aliases: []string{"member"},
		Short:   "Remove an account from an organization",
		Args:    cobra.ExactArgs(2),
	}
	command := &adminDeleteOrgMemberCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *adminDeleteOrgMemberCommand) run(_ *cobra.Command, args []string) {
	env, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	zrok, err := env.Client()
	if err != nil {
		panic(err)
	}

	req := admin.NewRemoveOrganizationMemberParams()
	req.Body.OrganizationToken = args[0]
	req.Body.Email = args[1]

	_, err = zrok.Admin.RemoveOrganizationMember(req, mustGetAdminAuth())
	if err != nil {
		panic(err)
	}

	logrus.Infof("removed '%v' from organization '%v", args[0], args[1])
}
