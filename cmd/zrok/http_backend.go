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
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok/tunnel"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/zrokdir"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func init() {
	httpCmd.AddCommand(newHttpBackendCommand().cmd)
}

type httpBackendCommand struct {
	quiet     bool
	basicAuth []string
	cmd       *cobra.Command
}

func newHttpBackendCommand() *httpBackendCommand {
	cmd := &cobra.Command{
		Use:     "backend <targetEndpoint>",
		Aliases: []string{"be"},
		Short:   "Create an HTTP binding",
		Args:    cobra.ExactArgs(1),
	}
	command := &httpBackendCommand{cmd: cmd}
	cmd.Flags().BoolVarP(&command.quiet, "quiet", "q", false, "Disable TUI 'chrome' for quiet operation")
	cmd.Flags().StringArrayVar(&command.basicAuth, "basic-auth", []string{}, "Basic authentication users (<username:password>,...")
	cmd.Run = command.run
	return command
}

func (self *httpBackendCommand) run(_ *cobra.Command, args []string) {
	if !self.quiet {
		if err := ui.Init(); err != nil {
			panic(err)
		}
		defer ui.Close()
		tb.SetInputMode(tb.InputEsc)
	}

	env, err := zrokdir.LoadEnvironment()
	if err != nil {
		panic(err)
	}
	zif, err := zrokdir.ZitiIdentityFile("environment")
	if err != nil {
		panic(err)
	}
	cfg := &backend.Config{
		IdentityPath:    zif,
		EndpointAddress: args[0],
	}

	zrok := newZrokClient()
	auth := httptransport.APIKeyAuth("X-TOKEN", "header", env.ZrokToken)
	req := tunnel.NewTunnelParams()
	req.Body = &rest_model_zrok.TunnelRequest{
		ZitiIdentityID: env.ZitiIdentityId,
		Endpoint:       cfg.EndpointAddress,
		AuthScheme:     string(model.None),
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
	resp, err := zrok.Tunnel.Tunnel(req, auth)
	if err != nil {
		panic(err)
	}
	cfg.Service = resp.Payload.Service

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		self.destroy(env.ZitiIdentityId, cfg, zrok, auth)
		os.Exit(0)
	}()

	httpProxy, err := backend.NewHTTP(cfg)
	if err != nil {
		panic(err)
	}

	go func() {
		if err := httpProxy.Run(); err != nil {
			panic(err)
		}
	}()

	if !self.quiet {
		ui.Clear()
		w, h := ui.TerminalDimensions()

		p := widgets.NewParagraph()
		p.Border = true
		p.Title = " access your zrok service "
		p.Text = fmt.Sprintf("%v%v", strings.Repeat(" ", (((w-12)-len(resp.Payload.ProxyEndpoint))/2)-1), resp.Payload.ProxyEndpoint)
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
						self.destroy(env.ZitiIdentityId, cfg, zrok, auth)
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
		logrus.Infof("access your zrok service: %v", resp.Payload.ProxyEndpoint)
		for {
			time.Sleep(30 * time.Second)
		}
	}
}

func (self *httpBackendCommand) destroy(id string, cfg *backend.Config, zrok *rest_client_zrok.Zrok, auth runtime.ClientAuthInfoWriter) {
	logrus.Infof("shutting down '%v'", cfg.Service)
	req := tunnel.NewUntunnelParams()
	req.Body = &rest_model_zrok.UntunnelRequest{
		ZitiIdentityID: id,
		Service:        cfg.Service,
	}
	if _, err := zrok.Tunnel.Untunnel(req, auth); err == nil {
		logrus.Infof("shutdown complete")
	} else {
		logrus.Errorf("error shutting down: %v", err)
	}
}
