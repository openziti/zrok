package main

import (
	"fmt"
	"os"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/rest_client_zrok/admin"
	"github.com/spf13/cobra"
)

func init() {
	adminCreateCmd.AddCommand(newAdminCreateIdentity().cmd)
}

type adminCreateIdentity struct {
	cmd *cobra.Command
}

func newAdminCreateIdentity() *adminCreateIdentity {
	cmd := &cobra.Command{
		Use:     "identity <name>",
		Aliases: []string{"id"},
		Short:   "Create an identity and policies for a public frontend",
		Args:    cobra.ExactArgs(1),
	}
	command := &adminCreateIdentity{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *adminCreateIdentity) run(_ *cobra.Command, args []string) {
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
		dl.Errorf("identity '%v' already exists at '%v'", name, zif)
		os.Exit(1)
	}

	zrok, err := env.Client()
	if err != nil {
		panic(err)
	}

	req := admin.NewCreateIdentityParams()
	req.Body.Name = name

	resp, err := zrok.Admin.CreateIdentity(req, mustGetAdminAuth())
	if err != nil {
		panic(err)
	}

	if err := env.SaveZitiIdentityNamed(name, resp.Payload.Cfg); err != nil {
		panic(err)
	}

	fmt.Printf("zrok identity '%v' created with ziti id '%v'\n", name, resp.Payload.Identity)
}
