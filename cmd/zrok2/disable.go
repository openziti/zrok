package main

import (
	"fmt"

	httpTransport "github.com/go-openapi/runtime/client"
	"github.com/openziti/zrok/v2/environment"
	restEnvironment "github.com/openziti/zrok/v2/rest_client_zrok/environment"
	"github.com/openziti/zrok/v2/tui"
	"github.com/pkg/errors"
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
		Long: "Disable (and clean up) the enabled zrok environment. When the zrok controller does not complete the " +
			"request, the local environment is kept: a busy or unreachable controller can be retried, and when the " +
			"controller refuses the request and the environment no longer exists there, the local copy under " +
			"'~/.zrok2' can be removed by hand.\n\n" + exitCodesHelp,
		Args: cobra.NoArgs,
	}
	command := &disableCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *disableCommand) run(_ *cobra.Command, _ []string) {
	env, err := environment.LoadRoot()
	if err != nil {
		if !panicInstead {
			tui.Error("unable to load environment", err)
		}
		panic(err)
	}

	if !env.IsEnabled() {
		tui.Error("no environment found; nothing to disable!", nil)
	}

	zrok, err := env.Client()
	if err != nil {
		if !panicInstead {
			tui.Error("could not create zrok client", err)
		}
		panic(err)
	}
	auth := httpTransport.APIKeyAuth("X-TOKEN", "header", env.Environment().AccountToken)
	req := restEnvironment.NewDisableParams()
	req.Body.Identity = env.Environment().ZitiIdentity

	_, err = zrok.Environment.Disable(req, auth)
	if err != nil {
		// the local environment is kept on every failure, a 401 included: a revoked account token answers
		// 401 as well as an environment the controller no longer knows, and only the user can tell which
		var unauthorized *restEnvironment.DisableUnauthorized
		if errors.As(err, &unauthorized) {
			exitWithFailure("unable to disable environment; the zrok controller refused the request and the local environment is kept; "+
				"if this environment no longer exists on the controller, the local copy under '~/.zrok2' can be removed by hand", err)
		}
		exitWithFailure("unable to disable environment; the local environment is kept", err)
	}
	if err := env.DeleteEnvironment(); err != nil {
		if !panicInstead {
			tui.Error("error removing zrok environment", err)
		}
		panic(err)
	}
	if err := env.DeleteZitiIdentityNamed(env.EnvironmentIdentityName()); err != nil {
		if !panicInstead {
			tui.Error("error removing zrok backend identity", err)
		}
	}
	fmt.Println("zrok environment disabled...")
}
