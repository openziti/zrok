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
	"os"
	"os/signal"
	"syscall"
)

type accessPrivateCommand struct {
	cmd *cobra.Command
}

func newAccessPrivateCommand() *accessPrivateCommand {
	cmd := &cobra.Command{
		Use:   "private <serviceToken>",
		Short: "Create a private frontend to access a service",
		Args:  cobra.ExactArgs(1),
	}
	command := &accessPrivateCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *accessPrivateCommand) run(_ *cobra.Command, args []string) {
	svcName := args[0]

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
		SvcName: svcName,
		ZID:     env.ZId,
	}
	_, err = zrok.Service.Access(req, auth)
	if err != nil {
		if !panicInstead {
			showError("unable to access", err)
		}
		panic(err)
	}

	cfg := private_frontend.DefaultConfig("backend")

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cmd.destroy(env.ZId, svcName, zrok, auth)
		os.Exit(0)
	}()

	frontend, err := private_frontend.NewHTTP(cfg)
	if err != nil {
		if !panicInstead {
			showError("unable to create private frontend", err)
		}
		panic(err)
	}
	if err := frontend.Run(); err != nil {
		if !panicInstead {
			showError("unable to run frontend", err)
		}
	}
}

func (cmd *accessPrivateCommand) destroy(envZId, svcName string, zrok *rest_client_zrok.Zrok, auth runtime.ClientAuthInfoWriter) {
	logrus.Debugf("shutting down '%v'", svcName)
	req := service.NewUnaccessParams()
	req.Body = &rest_model_zrok.UnaccessRequest{
		SvcName: svcName,
		ZID:     envZId,
	}
	if _, err := zrok.Service.Unaccess(req, auth); err == nil {
		logrus.Debugf("shutdown complete")
	} else {
		logrus.Errorf("error shutting down: %v", err)
	}
}
