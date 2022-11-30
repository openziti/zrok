package main

import (
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti-test-kitchen/zrok/endpoints/private_frontend"
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
		Use:   "private <serviceToken>",
		Short: "Create a private frontend to access a service",
		Args:  cobra.ExactArgs(1),
	}
	command := &accessPrivateCommand{cmd: cmd}
	cmd.Run = command.run
	cmd.Flags().StringVarP(&command.bindAddress, "bind", "b", "0.0.0.0:9191", "The address to bind the private frontend")
	return command
}

func (cmd *accessPrivateCommand) run(_ *cobra.Command, args []string) {
	svcToken := args[0]

	endpointUrl, err := url.Parse("http://" + cmd.bindAddress)
	if err != nil {
		if !panicInstead {
			showError("invalid endpoint address", err)
		}
		panic(err)
	}

	env, err := zrokdir.LoadEnvironment()
	if err != nil {
		if !panicInstead {
			showError("unable to load environment; did you 'zrok enable'?", err)
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
	req := service.NewAccessParams()
	req.Body = &rest_model_zrok.AccessRequest{
		SvcToken: svcToken,
		EnvZID:   env.ZId,
	}
	accessResp, err := zrok.Service.Access(req, auth)
	if err != nil {
		if !panicInstead {
			showError("unable to access", err)
		}
		panic(err)
	}
	logrus.Infof("allocated frontend '%v'", accessResp.Payload.FrontendToken)

	cfg := private_frontend.DefaultConfig("backend")
	cfg.SvcToken = svcToken
	cfg.Address = cmd.bindAddress

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cmd.destroy(accessResp.Payload.FrontendToken, env.ZId, svcToken, zrok, auth)
		os.Exit(0)
	}()

	frontend, err := private_frontend.NewHTTP(cfg)
	if err != nil {
		if !panicInstead {
			showError("unable to create private frontend", err)
		}
		panic(err)
	}

	logrus.Infof("access your service at: %v", endpointUrl.String())

	if err := frontend.Run(); err != nil {
		if !panicInstead {
			showError("unable to run frontend", err)
		}
	}
}

func (cmd *accessPrivateCommand) destroy(frotendName, envZId, svcToken string, zrok *rest_client_zrok.Zrok, auth runtime.ClientAuthInfoWriter) {
	logrus.Debugf("shutting down '%v'", svcToken)
	req := service.NewUnaccessParams()
	req.Body = &rest_model_zrok.UnaccessRequest{
		FrontendToken: frotendName,
		SvcToken:      svcToken,
		EnvZID:        envZId,
	}
	if _, err := zrok.Service.Unaccess(req, auth); err == nil {
		logrus.Debugf("shutdown complete")
	} else {
		logrus.Errorf("error shutting down: %v", err)
	}
}
