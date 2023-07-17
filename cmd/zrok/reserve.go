package main

import (
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/model"
	"github.com/openziti/zrok/rest_client_zrok/share"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/sdk"
	"github.com/openziti/zrok/tui"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

func init() {
	rootCmd.AddCommand(newReserveCommand().cmd)
}

type reserveCommand struct {
	basicAuth         []string
	frontendSelection []string
	backendMode       string
	cmd               *cobra.Command
}

func newReserveCommand() *reserveCommand {
	cmd := &cobra.Command{
		Use:   "reserve <public|private> <target>",
		Short: "Create a reserved share",
		Args:  cobra.ExactArgs(2),
	}
	command := &reserveCommand{cmd: cmd}
	cmd.Flags().StringArrayVar(&command.basicAuth, "basic-auth", []string{}, "Basic authentication users (<username:password>,...)")
	cmd.Flags().StringArrayVar(&command.frontendSelection, "frontends", []string{"public"}, "Selected frontends to use for the share")
	cmd.Flags().StringVar(&command.backendMode, "backend-mode", "proxy", "The backend mode {proxy, web, <tcpTunnel, udpTunnel>}")
	cmd.Run = command.run
	return command
}

func (cmd *reserveCommand) run(_ *cobra.Command, args []string) {
	shareMode := sdk.ShareMode(args[0])
	if shareMode != sdk.PublicShareMode && shareMode != sdk.PrivateShareMode {
		tui.Error("invalid sharing mode; expecting 'public' or 'private'", nil)
	}

	var target string
	switch cmd.backendMode {
	case "proxy":
		v, err := parseUrl(args[1])
		if err != nil {
			if !panicInstead {
				tui.Error("invalid target endpoint URL", err)
			}
			panic(err)
		}
		target = v

	case "web":
		target = args[1]
	}

	env, err := environment.LoadRoot()
	if err != nil {
		if !panicInstead {
			tui.Error("error loading environment", err)
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
	auth := httptransport.APIKeyAuth("X-TOKEN", "header", env.Environment().Token)
	req := share.NewShareParams()
	req.Body = &rest_model_zrok.ShareRequest{
		EnvZID:               env.Environment().ZitiIdentity,
		ShareMode:            string(shareMode),
		BackendMode:          cmd.backendMode,
		BackendProxyEndpoint: target,
		AuthScheme:           string(model.None),
		Reserved:             true,
	}
	if shareMode == sdk.PublicShareMode {
		req.Body.FrontendSelection = cmd.frontendSelection
	}
	if len(cmd.basicAuth) > 0 {
		logrus.Infof("configuring basic auth")
		req.Body.AuthScheme = string(model.Basic)
		for _, pair := range cmd.basicAuth {
			tokens := strings.Split(pair, ":")
			if len(tokens) == 2 {
				req.Body.AuthUsers = append(req.Body.AuthUsers, &rest_model_zrok.AuthUser{Username: strings.TrimSpace(tokens[0]), Password: strings.TrimSpace(tokens[1])})
			} else {
				panic(errors.Errorf("invalid username:password pair '%v'", pair))
			}
		}
	}

	resp, err := zrok.Share.Share(req, auth)
	if err != nil {
		if !panicInstead {
			tui.Error("unable to create tunnel", err)
		}
		panic(err)
	}

	logrus.Infof("your reserved share token is '%v'", resp.Payload.ShrToken)
	for _, fpe := range resp.Payload.FrontendProxyEndpoints {
		logrus.Infof("reserved frontend endpoint: %v", fpe)
	}
}
