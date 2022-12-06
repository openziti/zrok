package main

import (
	ui "github.com/gizak/termui/v3"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti-test-kitchen/zrok/model"
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok/service"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/zrokdir"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/url"
	"strings"
)

func init() {
	rootCmd.AddCommand(newReserveCommand().cmd)
}

type reserveCommand struct {
	basicAuth []string
	cmd       *cobra.Command
}

func newReserveCommand() *reserveCommand {
	cmd := &cobra.Command{
		Use:   "reserve <public|private> <targetEndpoint>",
		Short: "Create a reserved service",
		Args:  cobra.ExactArgs(2),
	}
	command := &reserveCommand{cmd: cmd}
	cmd.Flags().StringArrayVar(&command.basicAuth, "basic-auth", []string{}, "Basic authentication users (<username:password>,...)")
	cmd.Run = command.run
	return command
}

func (cmd *reserveCommand) run(_ *cobra.Command, args []string) {
	shareMode := args[0]
	if shareMode != "public" && shareMode != "private" {
		showError("invalid sharing mode; expecting 'public' or 'private'", nil)
	}

	targetEndpoint, err := url.Parse(args[1])
	if err != nil {
		if !panicInstead {
			showError("invalid target endpoint URL", err)
		}
		panic(err)
	}
	if targetEndpoint.Scheme == "" {
		targetEndpoint.Scheme = "https"
	}

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
	req := service.NewShareParams()
	req.Body = &rest_model_zrok.ShareRequest{
		EnvZID:               env.ZId,
		ShareMode:            shareMode,
		BackendMode:          "proxy",
		BackendProxyEndpoint: targetEndpoint.String(),
		AuthScheme:           string(model.None),
		Reserved:             true,
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

	resp, err := zrok.Service.Share(req, auth)
	if err != nil {
		ui.Close()
		if !panicInstead {
			showError("unable to create tunnel", err)
		}
		panic(err)
	}

	logrus.Infof("your reserved service token is '%v'", resp.Payload.SvcToken)
	for _, fpe := range resp.Payload.FrontendProxyEndpoints {
		logrus.Infof("reserved frontend endpoint: %v", fpe)
	}
}
