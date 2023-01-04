package main

import (
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
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
	shareCmd.AddCommand(newSharePrivateCommand().cmd)
}

type sharePrivateCommand struct {
	basicAuth   []string
	backendMode string
	cmd         *cobra.Command
}

func newSharePrivateCommand() *sharePrivateCommand {
	cmd := &cobra.Command{
		Use:   "private <targetEndpoint>",
		Short: "Share a target endpoint privately",
		Args:  cobra.ExactArgs(1),
	}
	command := &sharePrivateCommand{cmd: cmd}
	cmd.Flags().StringArrayVar(&command.basicAuth, "basic-auth", []string{}, "Basic authentication users (<username:password>,...")
	cmd.Flags().StringVar(&command.backendMode, "backend-mode", "proxy", "The backend mode {proxy, web}")
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

	env, err := zrokdir.LoadEnvironment()
	if err != nil {
		if !panicInstead {
			showError("unable to load environment; did you 'zrok enable'?", err)
		}
		panic(err)
	}
	zif, err := zrokdir.ZitiIdentityFile("backend")
	if err != nil {
		if !panicInstead {
			showError("unable to load ziti identity configuration", err)
		}
		panic(err)
	}

	zrok, err := zrokdir.ZrokClient(env.ApiEndpoint)
	if err != nil {
		if !panicInstead {
			showError("unable to create zrok client", err)
		}
		panic(err)
	}

	auth := httptransport.APIKeyAuth("X-TOKEN", "header", env.Token)
	req := share.NewShareParams()
	req.Body = &rest_model_zrok.ShareRequest{
		EnvZID:               env.ZId,
		ShareMode:            "private",
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
			showError("unable to create share", err)
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

	switch cmd.backendMode {
	case "proxy":
		cfg := &proxyBackend.Config{
			IdentityPath:    zif,
			EndpointAddress: target,
			ShrToken:        resp.Payload.ShrToken,
		}
		_, err = cmd.proxyBackendMode(cfg)
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
		_, err = cmd.webBackendMode(cfg)
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

	logrus.Infof("share with others; they will use this command for access: 'zrok access private %v'", resp.Payload.ShrToken)

	for {
		time.Sleep(30 * time.Second)
	}
}

func (cmd *sharePrivateCommand) proxyBackendMode(cfg *proxyBackend.Config) (backendHandler, error) {
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

func (cmd *sharePrivateCommand) webBackendMode(cfg *webBackend.Config) (backendHandler, error) {
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
