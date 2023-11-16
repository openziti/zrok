package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti/zrok/endpoints"
	"github.com/openziti/zrok/endpoints/drive"
	"github.com/openziti/zrok/endpoints/proxy"
	"github.com/openziti/zrok/endpoints/tcpTunnel"
	"github.com/openziti/zrok/endpoints/udpTunnel"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/rest_client_zrok/metadata"
	"github.com/openziti/zrok/rest_client_zrok/share"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/openziti/zrok/tui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	shareCmd.AddCommand(newShareReservedCommand().cmd)
}

type shareReservedCommand struct {
	overrideEndpoint string
	headless         bool
	insecure         bool
	cmd              *cobra.Command
}

func newShareReservedCommand() *shareReservedCommand {
	cmd := &cobra.Command{
		Use:   "reserved <shareToken>",
		Short: "Start a backend for a reserved share",
	}
	command := &shareReservedCommand{cmd: cmd}
	cmd.Flags().StringVar(&command.overrideEndpoint, "override-endpoint", "", "Override the stored target endpoint with a replacement")
	cmd.Flags().BoolVar(&command.headless, "headless", false, "Disable TUI and run headless")
	cmd.Flags().BoolVar(&command.insecure, "insecure", false, "Enable insecure TLS certificate validation")
	cmd.Run = command.run
	return command
}

func (cmd *shareReservedCommand) run(_ *cobra.Command, args []string) {
	shrToken := args[0]
	var target string

	root, err := environment.LoadRoot()
	if err != nil {
		if !panicInstead {
			tui.Error("error loading environment", err)
		}
		panic(err)
	}

	if !root.IsEnabled() {
		tui.Error("unable to load environment; did you 'zrok enable'?", nil)
	}

	zrok, err := root.Client()
	if err != nil {
		if !panicInstead {
			tui.Error("unable to create zrok client", err)
		}
		panic(err)
	}
	auth := httptransport.APIKeyAuth("X-TOKEN", "header", root.Environment().Token)
	req := metadata.NewGetShareDetailParams()
	req.ShrToken = shrToken
	resp, err := zrok.Metadata.GetShareDetail(req, auth)
	if err != nil {
		if !panicInstead {
			tui.Error("unable to retrieve reserved share", err)
		}
		panic(err)
	}
	target = cmd.overrideEndpoint
	if target == "" {
		target = resp.Payload.BackendProxyEndpoint
	}
	if sdk.BackendMode(resp.Payload.BackendMode) == sdk.CaddyBackendMode {
		cmd.headless = true
	}

	zif, err := root.ZitiIdentityNamed(root.EnvironmentIdentityName())
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

	var shareDescription string
	switch resp.Payload.ShareMode {
	case string(sdk.PublicShareMode):
		shareDescription = resp.Payload.FrontendEndpoint
	case string(sdk.PrivateShareMode):
		shareDescription = fmt.Sprintf("access your share with: %v", tui.Code.Render(fmt.Sprintf("zrok access private %v", shrToken)))
	}

	mdl := newShareModel(shrToken, []string{shareDescription}, sdk.ShareMode(resp.Payload.ShareMode), sdk.BackendMode(resp.Payload.BackendMode))
	if !cmd.headless {
		proxy.SetCaddyLoggingWriter(mdl)
	}

	requests := make(chan *endpoints.Request, 1024)
	switch resp.Payload.BackendMode {
	case "proxy":
		cfg := &proxy.BackendConfig{
			IdentityPath:    zif,
			EndpointAddress: target,
			ShrToken:        shrToken,
			Insecure:        cmd.insecure,
			Requests:        requests,
		}

		be, err := proxy.NewBackend(cfg)
		if err != nil {
			if !panicInstead {
				tui.Error("unable to create proxy backend handler", err)
			}
			panic(err)
		}

		go func() {
			if err := be.Run(); err != nil {
				logrus.Errorf("error running http proxy backend: %v", err)
			}
		}()

	case "web":
		cfg := &proxy.CaddyWebBackendConfig{
			IdentityPath: zif,
			WebRoot:      target,
			ShrToken:     shrToken,
			Requests:     requests,
		}

		be, err := proxy.NewCaddyWebBackend(cfg)
		if err != nil {
			if !panicInstead {
				tui.Error("error creating web backend", err)
			}
			panic(err)
		}

		go func() {
			if err := be.Run(); err != nil {
				logrus.Errorf("error running http web backend: %v", err)
			}
		}()

	case "tcpTunnel":
		cfg := &tcpTunnel.BackendConfig{
			IdentityPath:    zif,
			EndpointAddress: target,
			ShrToken:        shrToken,
			RequestsChan:    requests,
		}

		be, err := tcpTunnel.NewBackend(cfg)
		if err != nil {
			if !panicInstead {
				tui.Error("error creating tcpTunnel backend", err)
			}
			panic(err)
		}

		go func() {
			if err := be.Run(); err != nil {
				logrus.Errorf("error running tcpTunnel backend: %v", err)
			}
		}()

	case "udpTunnel":
		cfg := &udpTunnel.BackendConfig{
			IdentityPath:    zif,
			EndpointAddress: target,
			ShrToken:        shrToken,
			RequestsChan:    requests,
		}

		be, err := udpTunnel.NewBackend(cfg)
		if err != nil {
			if !panicInstead {
				tui.Error("error creating udpTunnel backend", err)
			}
			panic(err)
		}

		go func() {
			if err := be.Run(); err != nil {
				logrus.Errorf("error running udpTunnel backend: %v", err)
			}
		}()

	case "caddy":
		cfg := &proxy.CaddyfileBackendConfig{
			CaddyfilePath: target,
			Shr:           &sdk.Share{Token: shrToken, FrontendEndpoints: []string{resp.Payload.FrontendEndpoint}},
			Requests:      requests,
		}

		be, err := proxy.NewCaddyfileBackend(cfg)
		if err != nil {
			if !panicInstead {
				tui.Error("error creating caddy backend", err)
			}
			panic(err)
		}

		go func() {
			if err := be.Run(); err != nil {
				logrus.Errorf("error running caddy backend: %v", err)
			}
		}()

	case "drive":
		cfg := &drive.BackendConfig{
			IdentityPath: zif,
			DriveRoot:    target,
			ShrToken:     shrToken,
			Requests:     requests,
		}

		be, err := drive.NewBackend(cfg)
		if err != nil {
			if !panicInstead {
				tui.Error("error creating drive backend", err)
			}
			panic(err)
		}

		go func() {
			if err := be.Run(); err != nil {
				logrus.Errorf("error running drive backend: %v", err)
			}
		}()

	default:
		tui.Error("invalid backend mode", nil)
	}

	if cmd.headless {
		switch resp.Payload.ShareMode {
		case string(sdk.PublicShareMode):
			logrus.Infof("access your zrok share: %v", resp.Payload.FrontendEndpoint)

		case string(sdk.PrivateShareMode):
			logrus.Infof("use this command to access your zrok share: 'zrok access private %v'", shrToken)
		}
		for {
			select {
			case req := <-requests:
				logrus.Infof("%v -> %v %v", req.RemoteAddr, req.Method, req.Path)
			}
		}
	} else {
		logrus.SetOutput(mdl)
		prg := tea.NewProgram(mdl, tea.WithAltScreen())
		mdl.prg = prg

		go func() {
			for {
				select {
				case req := <-requests:
					prg.Send(req)
				}
			}
		}()

		if _, err := prg.Run(); err != nil {
			tui.Error("An error occurred", err)
		}

		close(requests)
	}
}
