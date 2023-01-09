package main

import (
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti-test-kitchen/zrok/endpoints/proxyBackend"
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok/metadata"
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok/share"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/tui"
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
		Use:   "reserved <shareToken>",
		Short: "Start a backend for a reserved share",
	}
	command := &shareReservedCommand{cmd: cmd}
	cmd.Flags().StringVar(&command.overrideEndpoint, "override-endpoint", "", "Override the stored target endpoint with a replacement")
	cmd.Run = command.run
	return command
}

func (cmd *shareReservedCommand) run(_ *cobra.Command, args []string) {
	shrToken := args[0]
	targetEndpoint := ""
	if cmd.overrideEndpoint != "" {
		e, err := url.Parse(cmd.overrideEndpoint)
		if err != nil {
			if !panicInstead {
				tui.Error("invalid override endpoint URL", err)
			}
			panic(err)
		}
		if e.Scheme == "" {
			e.Scheme = "https"
		}
		targetEndpoint = e.String()
	}

	zrd, err := zrokdir.Load()
	if err != nil {
		if !panicInstead {
			tui.Error("error loading zrokdir", err)
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
	req := metadata.NewGetShareDetailParams()
	req.ShrToken = shrToken
	resp, err := zrok.Metadata.GetShareDetail(req, auth)
	if err != nil {
		if !panicInstead {
			tui.Error("unable to retrieve reserved share", err)
		}
		panic(err)
	}
	if targetEndpoint == "" {
		targetEndpoint = resp.Payload.BackendProxyEndpoint
	}

	zif, err := zrokdir.ZitiIdentityFile("backend")
	if err != nil {
		if !panicInstead {
			tui.Error("unable to load ziti identity configuration", err)
		}
		panic(err)
	}
	cfg := &proxyBackend.Config{
		IdentityPath:    zif,
		EndpointAddress: targetEndpoint,
		ShrToken:        shrToken,
	}
	logrus.Infof("sharing target endpoint: '%v'", cfg.EndpointAddress)

	if resp.Payload.BackendProxyEndpoint != targetEndpoint {
		upReq := share.NewUpdateShareParams()
		upReq.Body = &rest_model_zrok.UpdateShareRequest{
			ShrToken:             shrToken,
			BackendProxyEndpoint: targetEndpoint,
		}
		if _, err := zrok.Share.UpdateShare(upReq, auth); err != nil {
			if !panicInstead {
				tui.Error("unable to update backend proxy endpoint", err)
			}
			panic(err)
		}
		logrus.Infof("updated backend proxy endpoint to: %v", targetEndpoint)
	} else {
		logrus.Infof("using existing backend proxy endpoint: %v", targetEndpoint)
	}

	httpProxy, err := proxyBackend.NewBackend(cfg)
	if err != nil {
		if !panicInstead {
			tui.Error("unable to create http backend", err)
		}
		panic(err)
	}

	go func() {
		if err := httpProxy.Run(); err != nil {
			if !panicInstead {
				tui.Error("unable to run http proxy", err)
			}
			panic(err)
		}
	}()

	switch resp.Payload.ShareMode {
	case "public":
		logrus.Infof("access your zrok share: %v", resp.Payload.FrontendEndpoint)

	case "private":
		logrus.Infof("use this command to access your zrok share: 'zrok access private %v'", shrToken)
	}

	for {
		time.Sleep(30 * time.Second)
	}
}
