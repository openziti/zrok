package main

import (
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti-test-kitchen/zrok/endpoints"
	"github.com/openziti-test-kitchen/zrok/endpoints/proxyBackend"
	"github.com/openziti-test-kitchen/zrok/endpoints/webBackend"
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok/metadata"
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok/share"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/tui"
	"github.com/openziti-test-kitchen/zrok/zrokdir"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
	var target string

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
	if target == "" {
		target = resp.Payload.BackendProxyEndpoint
	}

	zif, err := zrokdir.ZitiIdentityFile("backend")
	if err != nil {
		if !panicInstead {
			tui.Error("unable to load ziti identity configuration", err)
		}
		panic(err)
	}

	logrus.Infof("sharing target: '%v'", target)

	if resp.Payload.BackendProxyEndpoint != target {
		upReq := share.NewUpdateShareParams()
		upReq.Body = &rest_model_zrok.UpdateShareRequest{
			ShrToken:             shrToken,
			BackendProxyEndpoint: target,
		}
		if _, err := zrok.Share.UpdateShare(upReq, auth); err != nil {
			if !panicInstead {
				tui.Error("unable to update backend proxy endpoint", err)
			}
			panic(err)
		}
		logrus.Infof("updated backend proxy endpoint to: %v", target)
	} else {
		logrus.Infof("using existing backend proxy endpoint: %v", target)
	}

	switch resp.Payload.BackendMode {
	case "proxy":
		cfg := &proxyBackend.Config{
			IdentityPath:    zif,
			EndpointAddress: target,
			ShrToken:        shrToken,
		}
		_, err := cmd.proxyBackendMode(cfg)
		if err != nil {
			if !panicInstead {
				tui.Error("unable to create proxy backend handler", err)
			}
			panic(err)
		}

	case "web":
		cfg := &webBackend.Config{
			IdentityPath: zif,
			WebRoot:      target,
			ShrToken:     shrToken,
		}
		_, err := cmd.webBackendMode(cfg)
		if err != nil {
			if !panicInstead {
				tui.Error("unable to create web backend handler", err)
			}
			panic(err)
		}

	default:
		tui.Error("invalid backend mode", nil)
	}

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

func (cmd *shareReservedCommand) proxyBackendMode(cfg *proxyBackend.Config) (endpoints.BackendHandler, error) {
	be, err := proxyBackend.NewBackend(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "error creating http proxy backend")
	}

	go func() {
		if err := be.Run(); err != nil {
			logrus.Errorf("error running http proxy backend: %v", err)
		}
	}()

	return be, nil
}

func (cmd *shareReservedCommand) webBackendMode(cfg *webBackend.Config) (endpoints.BackendHandler, error) {
	be, err := webBackend.NewBackend(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "error creating http web backend")
	}

	go func() {
		if err := be.Run(); err != nil {
			logrus.Errorf("error running http web backend: %v", err)
		}
	}()

	return be, nil
}
