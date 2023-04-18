package main

import (
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti/zrok/endpoints/tcpTunnel"
	"github.com/openziti/zrok/rest_client_zrok/share"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/tui"
	"github.com/openziti/zrok/zrokdir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"time"
)

func init() {
	accessPrivateCmd.cmd.AddCommand(newAccessPrivateTunnelCommand().cmd)
}

type accessPrivateTunnelCommand struct {
	bindAddress string
	cmd         *cobra.Command
}

func newAccessPrivateTunnelCommand() *accessPrivateTunnelCommand {
	cmd := &cobra.Command{
		Use:   "tunnel <shareToken>",
		Short: "Create a private tunnel frontend to access a share",
		Args:  cobra.ExactArgs(1),
	}
	command := &accessPrivateTunnelCommand{cmd: cmd}
	cmd.Flags().StringVarP(&command.bindAddress, "bind", "b", "tcp:127.0.0.1:9191", "The address to bind the private tunnel")
	cmd.Run = command.run
	return command
}

func (cmd *accessPrivateTunnelCommand) run(_ *cobra.Command, args []string) {
	zrd, err := zrokdir.Load()
	if err != nil {
		tui.Error("unable to load zrokdir", err)
	}

	if zrd.Env == nil {
		tui.Error("unable to load environment; did you 'zrok enable'?", nil)
	}

	zrok, err := zrd.Client()
	if err != nil {
		tui.Error("unable to create zrok client", err)
	}

	auth := httptransport.APIKeyAuth("X-TOKEN", "header", zrd.Env.Token)
	req := share.NewAccessParams()
	req.Body = &rest_model_zrok.AccessRequest{
		ShrToken: args[0],
		EnvZID:   zrd.Env.ZId,
	}
	accessResp, err := zrok.Share.Access(req, auth)
	if err != nil {
		if !panicInstead {
			tui.Error("unable to access", err)
		}
		panic(err)
	}
	logrus.Infof("allocated frontend '%v'", accessResp.Payload.FrontendToken)

	fe, err := tcpTunnel.NewFrontend(&tcpTunnel.FrontendConfig{
		BindAddress:  cmd.bindAddress,
		IdentityName: "backend",
		ShrToken:     args[0],
	})
	if err != nil {
		panic(err)
	}
	if err := fe.Run(); err != nil {
		panic(err)
	}
	for {
		time.Sleep(50)
	}
}
