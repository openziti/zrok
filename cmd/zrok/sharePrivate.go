package main

import (
	"fmt"
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
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func init() {
	shareCmd.AddCommand(newSharePrivateCommand().cmd)
}

type sharePrivateCommand struct {
	basicAuth   []string
	backendMode string
	headless    bool
	insecure    bool
	cmd         *cobra.Command
}

func newSharePrivateCommand() *sharePrivateCommand {
	cmd := &cobra.Command{
		Use:   "private <target>",
		Short: "Share a target resource privately",
		Args:  cobra.ExactArgs(1),
	}
	command := &sharePrivateCommand{cmd: cmd}
	cmd.Flags().StringArrayVar(&command.basicAuth, "basic-auth", []string{}, "Basic authentication users (<username:password>,...")
	cmd.Flags().StringVar(&command.backendMode, "backend-mode", "proxy", "The backend mode {proxy, web}")
	cmd.Flags().BoolVar(&command.headless, "headless", false, "Disable TUI and run headless")
	cmd.Flags().BoolVar(&command.insecure, "insecure", false, "Enable insecure TLS certificate validation for <target>")
	cmd.Run = command.run
	return command
}

func (cmd *sharePrivateCommand) run(_ *cobra.Command, args []string) {
	var target string

	switch cmd.backendMode {
	case "proxy":
		targetEndpoint, err := url.Parse(args[0])
		if err != nil {
			if !panicInstead {
				tui.Error("invalid target endpoint URL", err)
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
		ShareMode:            "private",
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
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
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
		logrus.Infof("allow other to access your share with the following command:\nzrok access private %v", resp.Payload.ShrToken)
		for {
			select {
			case req := <-requestsChan:
				logrus.Infof("%v -> %v %v", req.RemoteAddr, req.Method, req.Path)
			}
		}

	} else {
		shareDescription := fmt.Sprintf("access your share with: %v", tui.Code.Render(fmt.Sprintf("zrok access private %v", resp.Payload.ShrToken)))
		mdl := newShareModel(resp.Payload.ShrToken, []string{shareDescription}, "private", cmd.backendMode)
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

func (cmd *sharePrivateCommand) proxyBackendMode(cfg *proxyBackend.Config) (endpoints.RequestHandler, error) {
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

func (cmd *sharePrivateCommand) webBackendMode(cfg *webBackend.Config) (endpoints.RequestHandler, error) {
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

func (cmd *sharePrivateCommand) destroy(id string, shrToken string, zrok *rest_client_zrok.Zrok, auth runtime.ClientAuthInfoWriter) {
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
