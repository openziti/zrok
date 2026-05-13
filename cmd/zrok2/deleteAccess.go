package main

import (
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/rest_client_zrok/metadata"
	"github.com/openziti/zrok/v2/rest_client_zrok/share"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	deleteCmd.AddCommand(newDeleteAccessCommand().cmd)
}

type deleteAccessCommand struct {
	cmd    *cobra.Command
	envZId string
}

func newDeleteAccessCommand() *deleteAccessCommand {
	cmd := &cobra.Command{
		Use:   "access <frontendToken>",
		Short: "Delete an access frontend",
		Args:  cobra.ExactArgs(1),
	}
	command := &deleteAccessCommand{cmd: cmd}
	cmd.Flags().StringVar(&command.envZId, "envzid", "", "Override environment ziti identifier")
	cmd.Run = command.run
	return command
}

func (cmd *deleteAccessCommand) run(_ *cobra.Command, args []string) {
	env, auth := mustGetEnvironmentAuth()
	zrok, err := env.Client()
	if err != nil {
		dl.Fatal(err)
	}

	envZId := resolveAccessDeleteEnvZId(env.Environment().ZitiIdentity, cmd.envZId)

	listReq := metadata.NewListAccessesParams()
	listReq.EnvZID = &envZId

	resp, err := zrok.Metadata.ListAccesses(listReq, auth)
	if err != nil {
		dl.Fatal(err)
	}

	req, err := resolveUnaccessRequest(args[0], envZId, resp.Payload.Accesses)
	if err != nil {
		dl.Fatal(err)
	}

	if _, err := zrok.Share.Unaccess(req, auth); err != nil {
		dl.Fatal(err)
	}

	dl.Infof("deleted access '%v' from environment '%v'", args[0], envZId)
}

func resolveAccessDeleteEnvZId(currentEnvZId, overrideEnvZId string) string {
	if overrideEnvZId != "" {
		return overrideEnvZId
	}
	return currentEnvZId
}

func resolveUnaccessRequest(frontendToken, envZId string, accesses []*rest_model_zrok.AccessSummary) (*share.UnaccessParams, error) {
	for _, access := range accesses {
		if access == nil || access.FrontendToken != frontendToken {
			continue
		}
		if access.ShareToken == "" {
			return nil, errors.Errorf("access '%v' in environment '%v' has no associated share token", frontendToken, envZId)
		}

		req := share.NewUnaccessParams()
		req.Body.FrontendToken = frontendToken
		req.Body.ShareToken = access.ShareToken
		req.Body.EnvZID = envZId
		return req, nil
	}

	return nil, errors.Errorf("access '%v' not found in environment '%v'", frontendToken, envZId)
}
