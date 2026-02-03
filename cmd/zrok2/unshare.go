package main

import (
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/rest_client_zrok/share"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newUnshareCommand().cmd)
}

type unshareCommand struct {
	cmd    *cobra.Command
	envZId string
}

func newUnshareCommand() *unshareCommand {
	cmd := &cobra.Command{
		Use:   "unshare",
		Short: "Remove a share",
		Args:  cobra.ExactArgs(1),
	}
	command := &unshareCommand{cmd: cmd}
	cmd.Flags().StringVar(&command.envZId, "envzid", "", "Override environment ziti identifier")
	cmd.Run = command.run
	return command
}

func (cmd *unshareCommand) run(_ *cobra.Command, args []string) {
	env, auth := mustGetEnvironmentAuth()
	zrok, err := env.Client()
	if err != nil {
		dl.Fatal(err)
	}

	req := share.NewUnshareParams()
	req.Body.EnvZID = env.Environment().ZitiIdentity
	if cmd.envZId != "" {
		req.Body.EnvZID = cmd.envZId
	}
	req.Body.ShareToken = args[0]

	_, err = zrok.Share.Unshare(req, auth)
	if err != nil {
		dl.Fatal(err)
	}

	dl.Infof("removed share '%v' from environment '%v'", req.Body.ShareToken, req.Body.EnvZID)
}
