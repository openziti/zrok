package main

import (
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/rest_client_zrok/share"
	"github.com/openziti/zrok/tui"
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
	env, err := environment.LoadRoot()
	if err != nil {
		if !panicInstead {
			tui.Error("unable to load environment", err)
		}
		panic(err)
	}

	if !env.IsEnabled() {
		tui.Error("unable to load environment; did you 'zrok enable'?", nil)
	}

	zrok, err := env.Client()
	if err != nil {
		if !panicInstead {
			tui.Error("unable to create zrok client", err)
		}
		panic(err)
	}

	auth := httptransport.APIKeyAuth("X-TOKEN", "header", env.Environment().AccountToken)
	req := share.NewUnshareParams()
	req.Body.EnvZID = env.Environment().ZitiIdentity
	req.Body.ShareToken = shrToken
	req.Body.Reserved = true
	if _, err := zrok.Share.Unshare(req, auth); err != nil {
		if !panicInstead {
			tui.Error("error releasing share", err)
		}
		panic(err)
	}

	logrus.Infof("reserved share '%v' released", shrToken)
}
