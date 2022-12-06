package main

import (
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	tb "github.com/nsf/termbox-go"
	"github.com/openziti-test-kitchen/zrok/endpoints/backend"
	"github.com/openziti-test-kitchen/zrok/model"
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok/service"
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
	cmd               *cobra.Command
}

func newSharePublicCommand() *sharePublicCommand {
	cmd := &cobra.Command{
		Use:   "public <targetEndpoint>",
		Short: "Share a target endpoint publicly",
		Args:  cobra.ExactArgs(1),
	}
	command := &sharePublicCommand{cmd: cmd}
	cmd.Flags().BoolVarP(&command.quiet, "quiet", "q", false, "Disable TUI 'chrome' for quiet operation")
	cmd.Flags().StringArrayVar(&command.basicAuth, "basic-auth", []string{}, "Basic authentication users (<username:password>,...)")
	cmd.Flags().StringArrayVar(&command.frontendSelection, "frontends", []string{"public"}, "Selected frontends to use for the share")
	cmd.Run = command.run
	return command
}

func (self *sharePublicCommand) run(_ *cobra.Command, args []string) {
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

	if !self.quiet {
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
	cfg := &backend.Config{
		IdentityPath:    zif,
		EndpointAddress: targetEndpoint.String(),
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
		ShareMode:            "public",
		FrontendSelection:    self.frontendSelection,
		BackendMode:          "proxy",
		BackendProxyEndpoint: cfg.EndpointAddress,
		AuthScheme:           string(model.None),
	}
	if len(self.basicAuth) > 0 {
		logrus.Infof("configuring basic auth")
		req.Body.AuthScheme = string(model.Basic)
		for _, pair := range self.basicAuth {
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
	cfg.Service = resp.Payload.SvcToken

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		self.destroy(env.ZId, cfg, zrok, auth)
		os.Exit(0)
	}()

	httpProxy, err := backend.NewHTTP(cfg)
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

	if !self.quiet {
		ui.Clear()
		w, h := ui.TerminalDimensions()

		p := widgets.NewParagraph()
		p.Border = true
		p.Title = " access your zrok service "
		p.Text = fmt.Sprintf("%v%v", strings.Repeat(" ", (((w-12)-len(resp.Payload.FrontendProxyEndpoint))/2)-1), resp.Payload.FrontendProxyEndpoint)
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
						self.destroy(env.ZId, cfg, zrok, auth)
						os.Exit(0)
					}
				}

			case <-ticker:
				currentRequests := float64(httpProxy.Requests())
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
		logrus.Infof("access your zrok service: %v", resp.Payload.FrontendProxyEndpoint)
		for {
			time.Sleep(30 * time.Second)
		}
	}
}

func (self *sharePublicCommand) destroy(id string, cfg *backend.Config, zrok *rest_client_zrok.Zrok, auth runtime.ClientAuthInfoWriter) {
	logrus.Debugf("shutting down '%v'", cfg.Service)
	req := service.NewUnshareParams()
	req.Body = &rest_model_zrok.UnshareRequest{
		EnvZID:   id,
		SvcToken: cfg.Service,
	}
	if _, err := zrok.Service.Unshare(req, auth); err == nil {
		logrus.Debugf("shutdown complete")
	} else {
		logrus.Errorf("error shutting down: %v", err)
	}
}
