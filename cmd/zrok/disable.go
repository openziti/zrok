package main

import (
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok/identity"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/zrokdir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newDisableCommand().cmd)
}

type disableCommand struct {
	cmd *cobra.Command
}

func newDisableCommand() *disableCommand {
	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Disable (and clean up) the enabled zrok environment",
		Args:  cobra.NoArgs,
	}
	command := &disableCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *disableCommand) run(_ *cobra.Command, args []string) {
	env, err := zrokdir.LoadEnvironment()
	if err != nil {
		panic(err)
	}
	zrok := newZrokClient()
	auth := httptransport.APIKeyAuth("X-TOKEN", "header", env.ZrokToken)
	req := identity.NewDisableParams()
	req.Body = &rest_model_zrok.DisableRequest{
		Identity: env.ZitiIdentityId,
	}
	_, err = zrok.Identity.Disable(req, auth)
	if err != nil {
		panic(err)
	}
	if err := zrokdir.Delete(); err != nil {
		panic(err)
	}
	logrus.Infof("environment disabled")
}
