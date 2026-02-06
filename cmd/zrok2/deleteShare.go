package main

import (
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/rest_client_zrok/share"
	"github.com/spf13/cobra"
)

func init() {
	deleteCmd.AddCommand(newDeleteShareCommand().cmd)
}

type deleteShareCommand struct {
	cmd    *cobra.Command
	envZId string
}

func newDeleteShareCommand() *deleteShareCommand {
	cmd := &cobra.Command{
		Use:   "share <shareToken>",
		Short: "Delete a share",
		Args:  cobra.ExactArgs(1),
	}
	command := &deleteShareCommand{cmd: cmd}
	cmd.Flags().StringVar(&command.envZId, "envzid", "", "Override environment ziti identifier")
	cmd.Run = command.run
	return command
}

func (cmd *deleteShareCommand) run(_ *cobra.Command, args []string) {
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

	dl.Infof("deleted share '%v' from environment '%v'", req.Body.ShareToken, req.Body.EnvZID)
}
