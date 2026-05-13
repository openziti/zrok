package main

import (
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/environment"
	"github.com/openziti/zrok/v2/rest_client_zrok/admin"
	"github.com/spf13/cobra"
)

func init() {
	adminCreateCmd.AddCommand(newAdminCreateSkipInterstitialGrantCommand().cmd)
}

type adminCreateSkipInterstitialGrantCommand struct {
	cmd *cobra.Command
}

func newAdminCreateSkipInterstitialGrantCommand() *adminCreateSkipInterstitialGrantCommand {
	cmd := &cobra.Command{
		Use:     "skip-interstitial-grant <email>",
		Aliases: []string{"sig"},
		Short:   "Grant skip interstitial to the specified account",
		Args:    cobra.ExactArgs(1),
	}
	command := &adminCreateSkipInterstitialGrantCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *adminCreateSkipInterstitialGrantCommand) run(_ *cobra.Command, args []string) {
	env, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	zrok, err := env.Client()
	if err != nil {
		panic(err)
	}

	req := admin.NewGrantSkipInterstitialParams()
	req.Body.Email = args[0]

	_, err = zrok.Admin.GrantSkipInterstitial(req, mustGetAdminAuth())
	if err != nil {
		panic(err)
	}

	dl.Infof("granted skip interstitial to '%v'", args[0])
}
