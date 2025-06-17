package main

import (
	"os"

	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/environment/env_core"
	"github.com/openziti/zrok/rest_client_zrok"
	"github.com/openziti/zrok/rest_client_zrok/admin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	adminCreateCmd.AddCommand(newAdminCreateSecretsIdentity().cmd)
}

type adminCreateSecretsIdentity struct {
	cmd *cobra.Command
}

func newAdminCreateSecretsIdentity() *adminCreateSecretsIdentity {
	cmd := &cobra.Command{
		Use:     "secrets-identity <name>",
		Aliases: []string{"si"},
		Short:   "Create a secrets identity for accessing the secrets listener",
		Args:    cobra.ExactArgs(1),
	}
	command := &adminCreateSecretsIdentity{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *adminCreateSecretsIdentity) run(_ *cobra.Command, args []string) {
	name := args[0]

	env, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}
	zif, err := env.ZitiIdentityNamed(name)
	if err != nil {
		panic(err)
	}
	if _, err := os.Stat(zif); err == nil {
		logrus.Errorf("identity '%v' already exists at '%v'", name, zif)
		os.Exit(1)
	}

	zrok, err := env.Client()
	if err != nil {
		panic(err)
	}

	zId, err := cmd.createIdentity(name, env, zrok)
	if err != nil {
		panic(err)
	}
	logrus.Infof("created identity '%v' with ziti id '%v'", name, zId)
}

func (cmd *adminCreateSecretsIdentity) createIdentity(name string, env env_core.Root, zrok *rest_client_zrok.Zrok) (zId string, err error) {
	req := admin.NewCreateIdentityParams()
	req.Body.Name = name

	resp, err := zrok.Admin.CreateIdentity(req, mustGetAdminAuth())
	if err != nil {
		return "", err
	}

	if err := env.SaveZitiIdentityNamed(name, resp.Payload.Cfg); err != nil {
		return "", err
	}

	return resp.Payload.Identity, nil
}
