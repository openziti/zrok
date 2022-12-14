package main

import (
	ui "github.com/gizak/termui/v3"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti-test-kitchen/zrok/endpoints/proxyBackend"
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok/service"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/zrokdir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/url"
	"time"
)

func init() {
	shareCmd.AddCommand(newShareReservedCommand().cmd)
}

type shareReservedCommand struct {
	overrideEndpoint string
	cmd              *cobra.Command
}

func newShareReservedCommand() *shareReservedCommand {
	cmd := &cobra.Command{
		Use:   "reserved <serviceToken>",
		Short: "Start a backend for a reserved service",
	}
	command := &shareReservedCommand{cmd: cmd}
	cmd.Flags().StringVar(&command.overrideEndpoint, "override-endpoint", "", "Override the stored target endpoint with a replacement")
	cmd.Run = command.run
	return command
}

func (cmd *shareReservedCommand) run(_ *cobra.Command, args []string) {
	svcToken := args[0]
	targetEndpoint := ""
	if cmd.overrideEndpoint != "" {
		e, err := url.Parse(cmd.overrideEndpoint)
		if err != nil {
			if !panicInstead {
				showError("invalid override endpoint URL", err)
			}
			panic(err)
		}
		if e.Scheme == "" {
			e.Scheme = "https"
		}
		targetEndpoint = e.String()
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
	req := service.NewGetServiceParams()
	req.Body = &rest_model_zrok.ServiceRequest{
		EnvZID:   env.ZId,
		SvcToken: svcToken,
	}
	resp, err := zrok.Service.GetService(req, auth)
	if err != nil {
		if !panicInstead {
			showError("unable to retrieve reserved service", err)
		}
		panic(err)
	}
	if targetEndpoint == "" {
		targetEndpoint = resp.Payload.BackendProxyEndpoint
	}

	zif, err := zrokdir.ZitiIdentityFile("backend")
	if err != nil {
		ui.Close()
		if !panicInstead {
			showError("unable to load ziti identity configuration", err)
		}
		panic(err)
	}
	cfg := &proxyBackend.Config{
		IdentityPath:    zif,
		EndpointAddress: targetEndpoint,
		Service:         svcToken,
	}
	logrus.Infof("sharing target endpoint: '%v'", cfg.EndpointAddress)

	if resp.Payload.BackendProxyEndpoint != targetEndpoint {
		upReq := service.NewUpdateShareParams()
		upReq.Body = &rest_model_zrok.UpdateShareRequest{
			ServiceToken:         svcToken,
			BackendProxyEndpoint: targetEndpoint,
		}
		if _, err := zrok.Service.UpdateShare(upReq, auth); err != nil {
			if !panicInstead {
				showError("unable to update backend proxy endpoint", err)
			}
			panic(err)
		}
		logrus.Infof("updated backend proxy endpoint to: %v", targetEndpoint)
	} else {
		logrus.Infof("using existing backend proxy endpoint: %v", targetEndpoint)
	}

	httpProxy, err := proxyBackend.NewBackend(cfg)
	if err != nil {
		ui.Close()
		if !panicInstead {
			showError("unable to create http backend", err)
		}
		panic(err)
	}

	go func() {
		if err := httpProxy.Run(); err != nil {
			if !panicInstead {
				showError("unable to run http proxy", err)
			}
			panic(err)
		}
	}()

	switch resp.Payload.ShareMode {
	case "public":
		logrus.Infof("access your zrok service: %v", resp.Payload.FrontendEndpoint)

	case "private":
		logrus.Infof("use this command to access your zrok service: 'zrok access private %v'", svcToken)
	}

	for {
		time.Sleep(30 * time.Second)
	}
}
