package main

import (
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/environment"
	"github.com/openziti/zrok/v2/rest_client_zrok/admin"
	"github.com/spf13/cobra"
)

func init() {
	adminDeleteCmd.AddCommand(newAdminDeleteSkipInterstitialGrantCommand().cmd)
}

type adminDeleteSkipInterstitialGrantCommand struct {
	cmd *cobra.Command
}

func newAdminDeleteSkipInterstitialGrantCommand() *adminDeleteSkipInterstitialGrantCommand {
	cmd := &cobra.Command{
		Use:     "skip-interstitial-grant <email>",
		Aliases: []string{"sig"},
		Short:   "Revoke skip interstitial from the specified account",
		Args:    cobra.ExactArgs(1),
	}
	command := &adminDeleteSkipInterstitialGrantCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *adminDeleteSkipInterstitialGrantCommand) run(_ *cobra.Command, args []string) {
	env, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	zrok, err := env.Client()
	if err != nil {
		panic(err)
	}

	req := admin.NewRevokeSkipInterstitialParams()
	req.Body.Email = args[0]

	_, err = zrok.Admin.RevokeSkipInterstitial(req, mustGetAdminAuth())
	if err != nil {
		panic(err)
	}

	dl.Infof("revoked skip interstitial from '%v'", args[0])
}
