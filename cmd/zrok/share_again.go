package main

import (
	ui "github.com/gizak/termui/v3"
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti-test-kitchen/zrok/endpoints/backend"
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok/service"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/zrokdir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func init() {
	shareCmd.AddCommand(newShareAgainCommand().cmd)
}

type shareAgainCommand struct {
	cmd *cobra.Command
}

func newShareAgainCommand() *shareAgainCommand {
	cmd := &cobra.Command{
		Use:   "again <serviceToken> <targetEndpoint>",
		Short: "Share a previously reserved service again",
		Args:  cobra.ExactArgs(2),
	}
	command := &shareAgainCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *shareAgainCommand) run(_ *cobra.Command, args []string) {
	targetEndpoint, err := url.Parse(args[1])
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
	svcToken := args[0]
	cfg := &backend.Config{
		IdentityPath:    zif,
		EndpointAddress: targetEndpoint.String(),
		Service:         svcToken,
	}

	zrok, err := zrokdir.ZrokClient(env.ApiEndpoint)
	if err != nil {
		if !panicInstead {
			showError("unable to create zrok client", err)
		}
		panic(err)
	}
	auth := httptransport.APIKeyAuth("X-TOKEN", "header", env.Token)

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

	logrus.Infof("share your zrok service; use this command for access: 'zrok access private %v'", svcToken)
	for {
		time.Sleep(30 * time.Second)
	}
}

func (self *shareAgainCommand) destroy(id string, cfg *backend.Config, zrok *rest_client_zrok.Zrok, auth runtime.ClientAuthInfoWriter) {
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
