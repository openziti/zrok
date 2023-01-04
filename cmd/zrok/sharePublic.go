package main

import (
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	tb "github.com/nsf/termbox-go"
	"github.com/openziti-test-kitchen/zrok/endpoints/proxyBackend"
	"github.com/openziti-test-kitchen/zrok/endpoints/webBackend"
	"github.com/openziti-test-kitchen/zrok/model"
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok/share"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/zrokdir"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func init() {
	shareCmd.AddCommand(newSharePublicCommand().cmd)
}

type sharePublicCommand struct {
	quiet             bool
	basicAuth         []string
	frontendSelection []string
	backendMode       string
	cmd               *cobra.Command
}

func newSharePublicCommand() *sharePublicCommand {
	cmd := &cobra.Command{
		Use:   "public <target>",
		Short: "Share a target resource publicly",
		Args:  cobra.ExactArgs(1),
	}
	command := &sharePublicCommand{cmd: cmd}
	cmd.Flags().BoolVarP(&command.quiet, "quiet", "q", false, "Disable TUI 'chrome' for quiet operation")
	cmd.Flags().StringArrayVar(&command.basicAuth, "basic-auth", []string{}, "Basic authentication users (<username:password>,...)")
	cmd.Flags().StringArrayVar(&command.frontendSelection, "frontends", []string{"public"}, "Selected frontends to use for the share")
	cmd.Flags().StringVar(&command.backendMode, "backend-mode", "proxy", "The backend mode {proxy, web}")
	cmd.Run = command.run
	return command
}

func (cmd *sharePublicCommand) run(_ *cobra.Command, args []string) {
	var target string

	switch cmd.backendMode {
	case "proxy":
		targetEndpoint, err := url.Parse(args[0])
		if err != nil {
			if !panicInstead {
				showError("invalid target endpoint URL", err)
			}
			panic(err)
		}
		if targetEndpoint.Scheme == "" {
			targetEndpoint.Scheme = "https"
		}
		target = targetEndpoint.String()

	case "web":
		target = args[0]

	default:
		showError(fmt.Sprintf("invalid backend mode '%v'; expected {proxy, web}", cmd.backendMode), nil)
	}

	if !cmd.quiet {
		if err := ui.Init(); err != nil {
			if !panicInstead {
				showError("unable to initialize user interface", err)
			}
			panic(err)
		}
		defer ui.Close()
		tb.SetInputMode(tb.InputEsc)
	}

	env, err := zrokdir.LoadEnvironment()
	if err != nil {
		ui.Close()
		if !panicInstead {
			showError("unable to load environment; did you 'zrok enable'?", err)
		}
		panic(err)
	}
	zif, err := zrokdir.ZitiIdentityFile("backend")
	if err != nil {
		ui.Close()
		if !panicInstead {
			showError("unable to load ziti identity configuration", err)
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
	req := share.NewShareParams()
	req.Body = &rest_model_zrok.ShareRequest{
		EnvZID:               env.ZId,
		ShareMode:            "public",
		FrontendSelection:    cmd.frontendSelection,
		BackendMode:          "proxy",
		BackendProxyEndpoint: target,
		AuthScheme:           string(model.None),
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
		ui.Close()
		if !panicInstead {
			showError("unable to create tunnel", err)
		}
		panic(err)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cmd.destroy(env.ZId, resp.Payload.ShrToken, zrok, auth)
		os.Exit(0)
	}()

	var bh backendHandler
	switch cmd.backendMode {
	case "proxy":
		cfg := &proxyBackend.Config{
			IdentityPath:    zif,
			EndpointAddress: target,
			ShrToken:        resp.Payload.ShrToken,
		}
		bh, err = cmd.proxyBackendMode(cfg)
		if err != nil {
			ui.Close()
			if !panicInstead {
				showError("unable to create proxy backend handler", err)
			}
			panic(err)
		}

	case "web":
		cfg := &webBackend.Config{
			IdentityPath: zif,
			WebRoot:      target,
			ShrToken:     resp.Payload.ShrToken,
		}
		bh, err = cmd.webBackendMode(cfg)
		if err != nil {
			ui.Close()
			if !panicInstead {
				showError("unable to create web backend handler", err)
			}
			panic(err)
		}

	default:
		ui.Close()
		showError("invalid backend mode", nil)
	}

	if !cmd.quiet {
		ui.Clear()
		w, h := ui.TerminalDimensions()

		p := widgets.NewParagraph()
		p.Border = true
		p.Title = " access your zrok share "
		p.Text = fmt.Sprintf("%v%v", strings.Repeat(" ", (((w-12)-len(resp.Payload.FrontendProxyEndpoints[0]))/2)-1), resp.Payload.FrontendProxyEndpoints[0])
		p.TextStyle = ui.Style{Fg: ui.ColorWhite}
		p.PaddingTop = 1
		p.SetRect(5, 5, w-10, 10)

		lastRequests := float64(0)
		var requestData []float64
		spk := widgets.NewSparkline()
		spk.Title = " requests "
		spk.Data = requestData
		spk.LineColor = ui.ColorCyan

		slg := widgets.NewSparklineGroup(spk)
		slg.SetRect(5, 11, w-10, h-5)

		ui.Render(p, slg)

		ticker := time.NewTicker(time.Second).C
		uiEvents := ui.PollEvents()
		for {
			select {
			case e := <-uiEvents:
				switch e.Type {
				case ui.ResizeEvent:
					ui.Clear()
					w, h = ui.TerminalDimensions()
					p.SetRect(5, 5, w-10, 10)
					slg.SetRect(5, 11, w-10, h-5)
					ui.Render(p, slg)

				case ui.KeyboardEvent:
					switch e.ID {
					case "q", "<C-c>":
						ui.Close()
						cmd.destroy(env.ZId, resp.Payload.ShrToken, zrok, auth)
						os.Exit(0)
					}
				}

			case <-ticker:
				currentRequests := float64(bh.Requests()())
				deltaRequests := currentRequests - lastRequests
				requestData = append(requestData, deltaRequests)
				lastRequests = currentRequests
				requestData = append(requestData, deltaRequests)
				for len(requestData) > w-17 {
					requestData = requestData[1:]
				}
				spk.Title = fmt.Sprintf(" requests (%d) ", int(currentRequests))
				spk.Data = requestData
				ui.Render(p, slg)
			}
		}
	} else {
		logrus.Infof("access your zrok share: %v", resp.Payload.FrontendProxyEndpoints[0])
		for {
			time.Sleep(30 * time.Second)
		}
	}
}

func (cmd *sharePublicCommand) proxyBackendMode(cfg *proxyBackend.Config) (backendHandler, error) {
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

func (cmd *sharePublicCommand) webBackendMode(cfg *webBackend.Config) (backendHandler, error) {
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

func (cmd *sharePublicCommand) destroy(id string, shrToken string, zrok *rest_client_zrok.Zrok, auth runtime.ClientAuthInfoWriter) {
	logrus.Debugf("shutting down '%v'", shrToken)
	req := share.NewUnshareParams()
	req.Body = &rest_model_zrok.UnshareRequest{
		EnvZID:   id,
		ShrToken: shrToken,
	}
	if _, err := zrok.Share.Unshare(req, auth); err == nil {
		logrus.Debugf("shutdown complete")
	} else {
		logrus.Errorf("error shutting down: %v", err)
	}
}
