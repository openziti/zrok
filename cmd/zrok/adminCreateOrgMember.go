package main

import (
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/rest_client_zrok/admin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	adminCreateCmd.AddCommand(newAdminCreateOrgMemberCommand().cmd)
}

type adminCreateOrgMemberCommand struct {
	cmd   *cobra.Command
	admin bool
}

func newAdminCreateOrgMemberCommand() *adminCreateOrgMemberCommand {
	cmd := &cobra.Command{
		Use:     "org-member <organizationToken> <accountEmail>",
		Aliases: []string{"member"},
		Short:   "Add an account to an organization",
		Args:    cobra.ExactArgs(2),
	}
	command := &adminCreateOrgMemberCommand{cmd: cmd}
	cmd.Flags().BoolVar(&command.admin, "admin", false, "Make the new account an admin of the organization")
	cmd.Run = command.run
	return command
}

func (cmd *adminCreateOrgMemberCommand) run(_ *cobra.Command, args []string) {
	env, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	zrok, err := env.Client()
	if err != nil {
		panic(err)
	}

	req := admin.NewAddOrganizationMemberParams()
	req.Body.OrganizationToken = args[0]
	req.Body.Email = args[1]
	req.Body.Admin = cmd.admin

	_, err = zrok.Admin.AddOrganizationMember(req, mustGetAdminAuth())
	if err != nil {
		panic(err)
	}

	logrus.Infof("added '%v' to organization '%v", args[0], args[1])
}
