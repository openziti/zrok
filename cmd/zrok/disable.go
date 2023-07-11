package main

import (
	"fmt"
	httpTransport "github.com/go-openapi/runtime/client"
	"github.com/openziti/zrok/environment/env_v0_3"
	restEnvironment "github.com/openziti/zrok/rest_client_zrok/environment"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/tui"
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

func (cmd *disableCommand) run(_ *cobra.Command, _ []string) {
	zrd, err := env_v0_3.Load()
	if err != nil {
		if !panicInstead {
			tui.Error("unable to load environment", err)
		}
		panic(err)
	}

	if zrd.Env == nil {
		tui.Error("no environment found; nothing to disable!", nil)
	}

	zrok, err := zrd.Client()
	if err != nil {
		if !panicInstead {
			tui.Error("could not create zrok client", err)
		}
		panic(err)
	}
	auth := httpTransport.APIKeyAuth("X-TOKEN", "header", zrd.Env.Token)
	req := restEnvironment.NewDisableParams()
	req.Body = &rest_model_zrok.DisableRequest{
		Identity: zrd.Env.ZId,
	}
	_, err = zrok.Environment.Disable(req, auth)
	if err != nil {
		logrus.Warnf("share cleanup failed (%v); will clean up local environment", err)
	}
	if err := env_v0_3.DeleteEnvironment(); err != nil {
		if !panicInstead {
			tui.Error("error removing zrok environment", err)
		}
		panic(err)
	}
	if err := env_v0_3.DeleteZitiIdentity("backend"); err != nil {
		if !panicInstead {
			tui.Error("error removing zrok backend identity", err)
		}
	}
	fmt.Println("zrok environment disabled...")
}
