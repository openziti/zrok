package main

import (
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/rest_client_zrok/share"
	"github.com/openziti/zrok/rest_model_zrok"
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
	zrd, err := environment.Load()
	if err != nil {
		if !panicInstead {
			tui.Error("unable to load environment", err)
		}
		panic(err)
	}

	if zrd.Env == nil {
		tui.Error("unable to load environment; did you 'zrok enable'?", nil)
	}

	zrok, err := zrd.Client()
	if err != nil {
		if !panicInstead {
			tui.Error("unable to create zrok client", err)
		}
		panic(err)
	}

	auth := httptransport.APIKeyAuth("X-TOKEN", "header", zrd.Env.Token)
	req := share.NewUnshareParams()
	req.Body = &rest_model_zrok.UnshareRequest{
		EnvZID:   zrd.Env.ZId,
		ShrToken: shrToken,
		Reserved: true,
	}
	if _, err := zrok.Share.Unshare(req, auth); err != nil {
		if !panicInstead {
			tui.Error("error releasing share", err)
		}
		panic(err)
	}

	logrus.Infof("reserved share '%v' released", shrToken)
}
