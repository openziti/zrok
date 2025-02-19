package main

import (
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/rest_client_zrok/admin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	adminCreateCmd.AddCommand(newAdminCreateOrganizationCommand().cmd)
}

type adminCreateOrganizationCommand struct {
	cmd         *cobra.Command
	description string
}

func newAdminCreateOrganizationCommand() *adminCreateOrganizationCommand {
	cmd := &cobra.Command{
		Use:     "organization",
		Aliases: []string{"org"},
		Short:   "Create a new organization",
		Args:    cobra.NoArgs,
	}
	command := &adminCreateOrganizationCommand{cmd: cmd}
	cmd.Flags().StringVarP(&command.description, "description", "d", "", "Organization description")
	cmd.Run = command.run
	return command
}

func (cmd *adminCreateOrganizationCommand) run(_ *cobra.Command, _ []string) {
	env, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	zrok, err := env.Client()
	if err != nil {
		panic(err)
	}

	req := admin.NewCreateOrganizationParams()
	req.Body = admin.CreateOrganizationBody{Description: cmd.description}

	resp, err := zrok.Admin.CreateOrganization(req, mustGetAdminAuth())
	if err != nil {
		panic(err)
	}

	logrus.Infof("created new organization with organization token '%v'", resp.Payload.OrganizationToken)
}
