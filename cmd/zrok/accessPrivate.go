package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti/zrok/endpoints"
	"github.com/openziti/zrok/endpoints/privateFrontend"
	"github.com/openziti/zrok/rest_client_zrok"
	"github.com/openziti/zrok/rest_client_zrok/share"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/tui"
	"github.com/openziti/zrok/zrokdir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/url"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	accessCmd.AddCommand(newAccessPrivateCommand().cmd)
}

type accessPrivateCommand struct {
	cmd         *cobra.Command
	bindAddress string
}

func newAccessPrivateCommand() *accessPrivateCommand {
	cmd := &cobra.Command{
		Use:   "private <shareToken>",
		Short: "Create a private frontend to access a share",
		Args:  cobra.ExactArgs(1),
	}
	command := &accessPrivateCommand{cmd: cmd}
	cmd.Run = command.run
	cmd.Flags().StringVarP(&command.bindAddress, "bind", "b", "127.0.0.1:9191", "The address to bind the private frontend")
	return command
}

func (cmd *accessPrivateCommand) run(_ *cobra.Command, args []string) {
	shrToken := args[0]

	endpointUrl, err := url.Parse("http://" + cmd.bindAddress)
	if err != nil {
		if !panicInstead {
			tui.Error("invalid endpoint address", err)
		}
		panic(err)
	}

	zrd, err := zrokdir.Load()
	if err != nil {
		tui.Error("unable to load zrokdir", err)
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

	cfg := privateFrontend.DefaultConfig("backend")
	cfg.ShrToken = shrToken
	cfg.Address = cmd.bindAddress
	cfg.RequestsChan = make(chan *endpoints.Request, 1024)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cmd.destroy(accessResp.Payload.FrontendToken, zrd.Env.ZId, shrToken, zrok, auth)
		os.Exit(0)
	}()

	frontend, err := privateFrontend.NewHTTP(cfg)
	if err != nil {
		if !panicInstead {
			tui.Error("unable to create private frontend", err)
		}
		panic(err)
	}

	go func() {
		if err := frontend.Run(); err != nil {
			if !panicInstead {
				tui.Error("unable to run frontend", err)
			}
		}
	}()

	mdl := newAccessModel(shrToken, endpointUrl.String())
	logrus.SetOutput(mdl)
	prg := tea.NewProgram(mdl, tea.WithAltScreen())
	mdl.prg = prg

	go func() {
		for {
			select {
			case req := <-cfg.RequestsChan:
				if req != nil {
					prg.Send(req)
				}
			}
		}
	}()

	if _, err := prg.Run(); err != nil {
		tui.Error("An error occurred", err)
	}

	close(cfg.RequestsChan)
	cmd.destroy(accessResp.Payload.FrontendToken, zrd.Env.ZId, shrToken, zrok, auth)
}

func (cmd *accessPrivateCommand) destroy(frotendName, envZId, shrToken string, zrok *rest_client_zrok.Zrok, auth runtime.ClientAuthInfoWriter) {
	logrus.Debugf("shutting down '%v'", shrToken)
	req := share.NewUnaccessParams()
	req.Body = &rest_model_zrok.UnaccessRequest{
		FrontendToken: frotendName,
		ShrToken:      shrToken,
		EnvZID:        envZId,
	}
	if _, err := zrok.Share.Unaccess(req, auth); err == nil {
		logrus.Debugf("shutdown complete")
	} else {
		logrus.Errorf("error shutting down: %v", err)
	}
}
