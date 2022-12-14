package main

import (
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok/admin"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/zrokdir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	adminUpdateCmd.AddCommand(newAdminUpdateFrontendCommand().cmd)
}

type adminUpdateFrontendCommand struct {
	cmd            *cobra.Command
	newPublicName  string
	newUrlTemplate string
}

func newAdminUpdateFrontendCommand() *adminUpdateFrontendCommand {
	cmd := &cobra.Command{
		Use:     "frontend <frontendToken>",
		Aliases: []string{"fe"},
		Short:   "Update a global public frontend",
		Args:    cobra.ExactArgs(1),
	}
	command := &adminUpdateFrontendCommand{cmd: cmd}
	cmd.Flags().StringVar(&command.newPublicName, "public-name", "", "Specify a new value for the public name")
	cmd.Flags().StringVar(&command.newUrlTemplate, "url-template", "", "Specify a new value for the url template")
	cmd.Run = command.run
	return command
}
func (cmd *adminUpdateFrontendCommand) run(_ *cobra.Command, args []string) {
	feToken := args[0]

	if cmd.newPublicName == "" && cmd.newUrlTemplate == "" {
		panic("must specify at least one of public name or url template")
	}

	zrok, err := zrokdir.ZrokClient(apiEndpoint)
	if err != nil {
		panic(err)
	}

	req := admin.NewUpdateFrontendParams()
	req.Body = &rest_model_zrok.UpdateFrontendRequest{
		FrontendToken: feToken,
		PublicName:    cmd.newPublicName,
		URLTemplate:   cmd.newUrlTemplate,
	}

	_, err = zrok.Admin.UpdateFrontend(req, mustGetAdminAuth())
	if err != nil {
		panic(err)
	}

	logrus.Infof("updated global frontend '%v'", feToken)
}
