package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti/zrok/endpoints"
	"github.com/openziti/zrok/endpoints/proxyBackend"
	"github.com/openziti/zrok/endpoints/webBackend"
	"github.com/openziti/zrok/model"
	"github.com/openziti/zrok/rest_client_zrok"
	"github.com/openziti/zrok/rest_client_zrok/share"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/tui"
	"github.com/openziti/zrok/zrokdir"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	shareCmd.AddCommand(newSharePublicCommand().cmd)
}

type sharePublicCommand struct {
	basicAuth         []string
	frontendSelection []string
	backendMode       string
	headless          bool
	insecure          bool
	cmd               *cobra.Command
}

func newSharePublicCommand() *sharePublicCommand {
	cmd := &cobra.Command{
		Use:   "public <target>",
		Short: "Share a target resource publicly",
		Args:  cobra.ExactArgs(1),
	}
	command := &sharePublicCommand{cmd: cmd}
	cmd.Flags().StringArrayVar(&command.basicAuth, "basic-auth", []string{}, "Basic authentication users (<username:password>,...)")
	cmd.Flags().StringArrayVar(&command.frontendSelection, "frontends", []string{"public"}, "Selected frontends to use for the share")
	cmd.Flags().StringVar(&command.backendMode, "backend-mode", "proxy", "The backend mode {proxy, web}")
	cmd.Flags().BoolVar(&command.headless, "headless", false, "Disable TUI and run headless")
	cmd.Flags().BoolVar(&command.insecure, "insecure", false, "Enable insecure TLS certificate validation for <target>")
	cmd.Run = command.run
	return command
}

func (cmd *sharePublicCommand) run(_ *cobra.Command, args []string) {
	var target string

	switch cmd.backendMode {
	case "proxy":
		v, err := parseUrl(args[0])
		if err != nil {
			if !panicInstead {
				tui.Error("invalid target endpoint URL", err)
			}
			panic(err)
		}
		target = v

	case "web":
		target = args[0]

	default:
		tui.Error(fmt.Sprintf("invalid backend mode '%v'; expected {proxy, web}", cmd.backendMode), nil)
	}

	zrd, err := zrokdir.Load()
	if err != nil {
		if !panicInstead {
			tui.Error("unable to load zrokdir", err)
		}
		panic(err)
	}

	if zrd.Env == nil {
		tui.Error("unable to load environment; did you 'zrok enable'?", nil)
	}

	zif, err := zrokdir.ZitiIdentityFile("backend")
	if err != nil {
		if !panicInstead {
			tui.Error("unable to load ziti identity configuration", err)
		}
		panic(err)
	}

	zrok, err := zrd.Client()
	if err != nil {
		if !panicInstead {
			tui.Error("unable to create zrok client", err)
		}
		panic(err)
	}

	auth := httptransport.APIKeyAuth("X-TOKEN", "header", zrd.Env.Token)
	req := share.NewShareParams()
	req.Body = &rest_model_zrok.ShareRequest{
		EnvZID:               zrd.Env.ZId,
		ShareMode:            "public",
		FrontendSelection:    cmd.frontendSelection,
		BackendMode:          cmd.backendMode,
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
		if !panicInstead {
			tui.Error("unable to create share", err)
		}
		panic(err)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		<-c
		cmd.destroy(zrd.Env.ZId, resp.Payload.ShrToken, zrok, auth)
		os.Exit(0)
	}()

	requestsChan := make(chan *endpoints.Request, 1024)
	switch cmd.backendMode {
	case "proxy":
		cfg := &proxyBackend.Config{
			IdentityPath:    zif,
			EndpointAddress: target,
			ShrToken:        resp.Payload.ShrToken,
			Insecure:        cmd.insecure,
			RequestsChan:    requestsChan,
		}
		_, err = cmd.proxyBackendMode(cfg)
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
			ShrToken:     resp.Payload.ShrToken,
			RequestsChan: requestsChan,
		}
		_, err = cmd.webBackendMode(cfg)
		if err != nil {
			if !panicInstead {
				tui.Error("unable to create web backend handler", err)
			}
			panic(err)
		}

	default:
		tui.Error("invalid backend mode", nil)
	}

	if cmd.headless {
		logrus.Infof("access your zrok share at the following endpoints:\n %v", strings.Join(resp.Payload.FrontendProxyEndpoints, "\n"))
		for {
			select {
			case req := <-requestsChan:
				logrus.Infof("%v -> %v %v", req.RemoteAddr, req.Method, req.Path)
			}
		}

	} else {
		mdl := newShareModel(resp.Payload.ShrToken, resp.Payload.FrontendProxyEndpoints, "public", cmd.backendMode)
		logrus.SetOutput(mdl)
		prg := tea.NewProgram(mdl, tea.WithAltScreen())
		mdl.prg = prg

		go func() {
			for {
				select {
				case req := <-requestsChan:
					prg.Send(req)
				}
			}
		}()

		if _, err := prg.Run(); err != nil {
			tui.Error("An error occurred", err)
		}

		close(requestsChan)
		cmd.destroy(zrd.Env.ZId, resp.Payload.ShrToken, zrok, auth)
	}
}

func (cmd *sharePublicCommand) proxyBackendMode(cfg *proxyBackend.Config) (endpoints.RequestHandler, error) {
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

func (cmd *sharePublicCommand) webBackendMode(cfg *webBackend.Config) (endpoints.RequestHandler, error) {
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
