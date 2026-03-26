package main

import (
	"fmt"

	"github.com/openziti/zrok/v2/environment"
	"github.com/openziti/zrok/v2/rest_client_zrok/admin"
	"github.com/spf13/cobra"
)

func init() {
	adminListCmd.AddCommand(newAdminListSkipInterstitialGrantCommand().cmd)
}

type adminListSkipInterstitialGrantCommand struct {
	cmd *cobra.Command
}

func newAdminListSkipInterstitialGrantCommand() *adminListSkipInterstitialGrantCommand {
	cmd := &cobra.Command{
		Use:     "skip-interstitial-grant <email>",
		Aliases: []string{"sig"},
		Short:   "Check if the specified account has the skip interstitial grant",
		Args:    cobra.ExactArgs(1),
	}
	command := &adminListSkipInterstitialGrantCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *adminListSkipInterstitialGrantCommand) run(_ *cobra.Command, args []string) {
	env, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	zrok, err := env.Client()
	if err != nil {
		panic(err)
	}

	req := admin.NewGetSkipInterstitialGrantParams()
	req.Email = args[0]

	resp, err := zrok.Admin.GetSkipInterstitialGrant(req, mustGetAdminAuth())
	if err != nil {
		panic(err)
	}

	fmt.Printf("account '%v' skip interstitial grant: %v\n", resp.Payload.Email, resp.Payload.Granted)
}
