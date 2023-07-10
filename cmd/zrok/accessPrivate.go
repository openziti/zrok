package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti/zrok/endpoints"
	"github.com/openziti/zrok/endpoints/proxy"
	"github.com/openziti/zrok/endpoints/tcpTunnel"
	"github.com/openziti/zrok/endpoints/udpTunnel"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/rest_client_zrok"
	"github.com/openziti/zrok/rest_client_zrok/share"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/tui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func init() {
	accessCmd.AddCommand(newAccessPrivateCommand().cmd)
}

type accessPrivateCommand struct {
	bindAddress string
	headless    bool
	cmd         *cobra.Command
}

func newAccessPrivateCommand() *accessPrivateCommand {
	cmd := &cobra.Command{
		Use:   "private <shareToken>",
		Short: "Create a private frontend to access a share",
		Args:  cobra.ExactArgs(1),
	}
	command := &accessPrivateCommand{cmd: cmd}
	cmd.Flags().BoolVar(&command.headless, "headless", false, "Disable TUI and run headless")
	cmd.Run = command.run
	cmd.Flags().StringVarP(&command.bindAddress, "bind", "b", "127.0.0.1:9191", "The address to bind the private frontend")
	return command
}

func (cmd *accessPrivateCommand) run(_ *cobra.Command, args []string) {
	shrToken := args[0]

	zrd, err := environment.Load()
	if err != nil {
		tui.Error("error loading environment", err)
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
	req := share.NewAccessParams()
	req.Body = &rest_model_zrok.AccessRequest{
		ShrToken: shrToken,
		EnvZID:   zrd.Env.ZId,
	}
	accessResp, err := zrok.Share.Access(req, auth)
	if err != nil {
		if !panicInstead {
			tui.Error("unable to access", err)
		}
		panic(err)
	}
	logrus.Infof("allocated frontend '%v'", accessResp.Payload.FrontendToken)

	protocol := "http://"
	switch accessResp.Payload.BackendMode {
	case "tcpTunnel":
		protocol = "tcp://"
	case "udpTunnel":
		protocol = "udp://"
	}

	endpointUrl, err := url.Parse(protocol + cmd.bindAddress)
	if err != nil {
		if !panicInstead {
			tui.Error("invalid endpoint address", err)
		}
		panic(err)
	}

	requests := make(chan *endpoints.Request, 1024)
	switch accessResp.Payload.BackendMode {
	case "tcpTunnel":
		fe, err := tcpTunnel.NewFrontend(&tcpTunnel.FrontendConfig{
			BindAddress:  cmd.bindAddress,
			IdentityName: "backend",
			ShrToken:     args[0],
			RequestsChan: requests,
		})
		if err != nil {
			if !panicInstead {
				tui.Error("unable to create private frontend", err)
			}
			panic(err)
		}
		go func() {
			if err := fe.Run(); err != nil {
				if !panicInstead {
					tui.Error("error starting frontend", err)
				}
				panic(err)
			}
		}()

	case "udpTunnel":
		fe, err := udpTunnel.NewFrontend(&udpTunnel.FrontendConfig{
			BindAddress:  cmd.bindAddress,
			IdentityName: "backend",
			ShrToken:     args[0],
			RequestsChan: requests,
			IdleTime:     time.Minute,
		})
		if err != nil {
			if !panicInstead {
				tui.Error("unable to create private frontend", err)
			}
			panic(err)
		}
		go func() {
			if err := fe.Run(); err != nil {
				if !panicInstead {
					tui.Error("error starting frontend", err)
				}
				panic(err)
			}
		}()

	default:
		cfg := proxy.DefaultFrontendConfig("backend")
		cfg.ShrToken = shrToken
		cfg.Address = cmd.bindAddress
		cfg.RequestsChan = requests
		fe, err := proxy.NewFrontend(cfg)
		if err != nil {
			if !panicInstead {
				tui.Error("unable to create private frontend", err)
			}
			panic(err)
		}
		go func() {
			if err := fe.Run(); err != nil {
				if !panicInstead {
					tui.Error("unable to run frontend", err)
				}
			}
		}()
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cmd.destroy(accessResp.Payload.FrontendToken, zrd.Env.ZId, shrToken, zrok, auth)
		os.Exit(0)
	}()

	if cmd.headless {
		logrus.Infof("access the zrok share at the followind endpoint: %v", endpointUrl.String())
		for {
			select {
			case req := <-requests:
				logrus.Infof("%v -> %v %v", req.RemoteAddr, req.Method, req.Path)
			}
		}

	} else {
		mdl := newAccessModel(shrToken, endpointUrl.String())
		logrus.SetOutput(mdl)
		prg := tea.NewProgram(mdl, tea.WithAltScreen())
		mdl.prg = prg

		go func() {
			for {
				select {
				case req := <-requests:
					if req != nil {
						prg.Send(req)
					}
				}
			}
		}()

		if _, err := prg.Run(); err != nil {
			tui.Error("An error occurred", err)
		}

		close(requests)
		cmd.destroy(accessResp.Payload.FrontendToken, zrd.Env.ZId, shrToken, zrok, auth)
	}
}

func (cmd *accessPrivateCommand) destroy(frontendName, envZId, shrToken string, zrok *rest_client_zrok.Zrok, auth runtime.ClientAuthInfoWriter) {
	logrus.Debugf("shutting down '%v'", shrToken)
	req := share.NewUnaccessParams()
	req.Body = &rest_model_zrok.UnaccessRequest{
		FrontendToken: frontendName,
		ShrToken:      shrToken,
		EnvZID:        envZId,
	}
	if _, err := zrok.Share.Unaccess(req, auth); err == nil {
		logrus.Debugf("shutdown complete")
	} else {
		logrus.Errorf("error shutting down: %v", err)
	}
}
