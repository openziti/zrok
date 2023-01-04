package main

import (
	ui "github.com/gizak/termui/v3"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok/share"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/zrokdir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newReleaseCommand().cmd)
}

type releaseCommand struct {
	cmd *cobra.Command
}

func newReleaseCommand() *releaseCommand {
	cmd := &cobra.Command{
		Use:   "release <shareToken>",
		Short: "Release a reserved share",
		Args:  cobra.ExactArgs(1),
	}
	command := &releaseCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *releaseCommand) run(_ *cobra.Command, args []string) {
	shrToken := args[0]
	env, err := zrokdir.LoadEnvironment()
	if err != nil {
		ui.Close()
		if !panicInstead {
			showError("unable to load environment; did you 'zrok enable'?", err)
		}
		panic(err)
	}

	zrok, err := zrokdir.ZrokClient(env.ApiEndpoint)
	if err != nil {
		ui.Close()
		if !panicInstead {
			showError("unable to create zrok client", err)
		}
		panic(err)
	}
	auth := httptransport.APIKeyAuth("X-TOKEN", "header", env.Token)
	req := share.NewUnshareParams()
	req.Body = &rest_model_zrok.UnshareRequest{
		EnvZID:   env.ZId,
		ShrToken: shrToken,
		Reserved: true,
	}
	if _, err := zrok.Share.Unshare(req, auth); err != nil {
		if !panicInstead {
			showError("error releasing share", err)
		}
		panic(err)
	}

	logrus.Infof("reserved share '%v' released", shrToken)
}
