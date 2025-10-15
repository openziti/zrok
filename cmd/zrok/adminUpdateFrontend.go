package main

import (
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/rest_client_zrok/admin"
	"github.com/spf13/cobra"
)

func init() {
	adminUpdateCmd.AddCommand(newAdminUpdateFrontendCommand().cmd)
}

type adminUpdateFrontendCommand struct {
	cmd            *cobra.Command
	newPublicName  string
	newUrlTemplate string
	dynamic        bool
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
	cmd.Flags().BoolVar(&command.dynamic, "dynamic", false, "Set dynamic mode for the frontend")
	cmd.Run = command.run
	return command
}
func (cmd *adminUpdateFrontendCommand) run(cobraCmd *cobra.Command, args []string) {
	feToken := args[0]

	dynamicSet := cobraCmd.Flags().Changed("dynamic")
	if cmd.newPublicName == "" && cmd.newUrlTemplate == "" && !dynamicSet {
		panic("must specify at least one of public name, url template, or dynamic")
	}

	env, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	zrok, err := env.Client()
	if err != nil {
		panic(err)
	}

	req := admin.NewUpdateFrontendParams()
	req.Body.FrontendToken = feToken
	req.Body.PublicName = cmd.newPublicName
	req.Body.URLTemplate = cmd.newUrlTemplate
	req.Body.Dynamic = cmd.dynamic
	req.Body.DynamicSet = dynamicSet

	_, err = zrok.Admin.UpdateFrontend(req, mustGetAdminAuth())
	if err != nil {
		panic(err)
	}

	dl.Infof("updated global frontend '%v'", feToken)
}
