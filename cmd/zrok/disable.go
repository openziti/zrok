package main

import (
	"fmt"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok/identity"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/zrokdir"
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
		if !panicInstead {
			showError("could not load environment; not active?", err)
		}
		panic(err)
	}
	zrok, err := zrokdir.ZrokClient(env.ApiEndpoint)
	if err != nil {
		if !panicInstead {
			showError("could not create zrok service client", err)
		}
		panic(err)
	}
	auth := httptransport.APIKeyAuth("X-TOKEN", "header", env.Token)
	req := identity.NewDisableParams()
	req.Body = &rest_model_zrok.DisableRequest{
		Identity: env.ZId,
	}
	_, err = zrok.Identity.Disable(req, auth)
	if err != nil {
		if !panicInstead {
			showError("zrok service call failed", err)
		}
		panic(err)
	}
	if err := zrokdir.DeleteEnvironment(); err != nil {
		if !panicInstead {
			showError("error removing zrok environment", err)
		}
		panic(err)
	}
	if err := zrokdir.DeleteZitiIdentity("backend"); err != nil {
		if !panicInstead {
			showError("error removing zrok backend identity", err)
		}
	}
	fmt.Printf("zrok environment '%v' disabled for '%v'\n", env.ZId, env.Token)
}
