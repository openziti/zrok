package main

import (
	ui "github.com/gizak/termui/v3"
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
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
	shareCmd.AddCommand(newSharePrivateCommand().cmd)
}

type sharePrivateCommand struct {
	basicAuth []string
	cmd       *cobra.Command
}

func newSharePrivateCommand() *sharePrivateCommand {
	cmd := &cobra.Command{
		Use:   "private <targetEndpoint>",
		Short: "Share a target endpoint privately",
		Args:  cobra.ExactArgs(1),
	}
	command := &sharePrivateCommand{cmd: cmd}
	cmd.Flags().StringArrayVar(&command.basicAuth, "basic-auth", []string{}, "Basic authentication users (<username:password>,...")
	cmd.Run = command.run
	return command
}

func (cmd *sharePrivateCommand) run(_ *cobra.Command, args []string) {
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
	cfg := &backend.Config{
		IdentityPath:    zif,
		EndpointAddress: targetEndpoint.String(),
	}

	zrok, err := zrokdir.ZrokClient(env.ApiEndpoint)
	if err != nil {
		if !panicInstead {
			showError("unable to create zrok client", err)
		}
		panic(err)
	}

	auth := httptransport.APIKeyAuth("X-TOKEN", "header", env.Token)
	req := service.NewShareParams()
	req.Body = &rest_model_zrok.ShareRequest{
		ZID:                  env.ZId,
		ShareMode:            "private",
		BackendMode:          "proxy",
		BackendProxyEndpoint: cfg.EndpointAddress,
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
	resp, err := zrok.Service.Share(req, auth)
	if err != nil {
		ui.Close()
		if !panicInstead {
			showError("unable to create share", err)
		}
		panic(err)
	}
	cfg.Service = resp.Payload.SvcName

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cmd.destroy(env.ZId, cfg, zrok, auth)
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

	logrus.Infof("share your zrok service; use this command for access: 'zrok access private %v'", resp.Payload.SvcName)

	for {
		time.Sleep(30 * time.Second)
	}
}

func (self *sharePrivateCommand) destroy(id string, cfg *backend.Config, zrok *rest_client_zrok.Zrok, auth runtime.ClientAuthInfoWriter) {
	logrus.Debugf("shutting down '%v'", cfg.Service)
	req := service.NewUnshareParams()
	req.Body = &rest_model_zrok.UnshareRequest{
		ZID:     id,
		SvcName: cfg.Service,
	}
	if _, err := zrok.Service.Unshare(req, auth); err == nil {
		logrus.Debugf("shutdown complete")
	} else {
		logrus.Errorf("error shutting down: %v", err)
	}
}
