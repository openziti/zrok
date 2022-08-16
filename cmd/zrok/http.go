package main

import (
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti-test-kitchen/zrok/http"
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
	rootCmd.AddCommand(newHttpCommand().cmd)
}

type httpCommand struct {
	basicAuth []string
	cmd       *cobra.Command
}

func newHttpCommand() *httpCommand {
	cmd := &cobra.Command{
		Use:   "http <endpoint>",
		Short: "Start an HTTP terminator",
		Args:  cobra.ExactArgs(1),
	}
	command := &httpCommand{cmd: cmd}
	cmd.Flags().StringArrayVar(&command.basicAuth, "basic-auth", []string{}, "Basic authentication users (<username:password>,...")
	cmd.Run = command.run
	return command
}

func (self *httpCommand) run(_ *cobra.Command, args []string) {
	/*
		if err := ui.Init(); err != nil {
			panic(err)
		}
		defer ui.Close()
		tb.SetInputMode(tb.InputEsc)
	*/

	idCfg, err := zrokdir.IdentityConfigFile()
	if err != nil {
		panic(err)
	}
	cfg := &http.Config{
		IdentityPath:    idCfg,
		EndpointAddress: args[0],
	}
	id, err := zrokdir.ReadIdentityId()
	if err != nil {
		panic(err)
	}
	token, err := zrokdir.ReadToken()
	if err != nil {
		panic(err)
	}

	zrok := newZrokClient()
	auth := httptransport.APIKeyAuth("X-TOKEN", "header", token)
	req := tunnel.NewTunnelParams()
	req.Body = &rest_model_zrok.TunnelRequest{
		ZitiIdentityID: id,
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
		cleanupHttp(id, cfg, zrok, auth)
		os.Exit(0)
	}()

	httpProxy, err := http.New(cfg)
	if err != nil {
		panic(err)
	}

	go func() {
		if err := httpProxy.Run(); err != nil {
			panic(err)
		}
	}()

	/*
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
						cleanupHttp(id, cfg, zrok, auth)
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
	*/
	for {
		time.Sleep(30 * time.Second)
	}
}

func cleanupHttp(id string, cfg *http.Config, zrok *rest_client_zrok.Zrok, auth runtime.ClientAuthInfoWriter) {
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
