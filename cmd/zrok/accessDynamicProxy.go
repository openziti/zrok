package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/endpoints/dynamicProxy"
	"github.com/openziti/zrok/v2/environment"
	"github.com/openziti/zrok/v2/tui"
	"github.com/spf13/cobra"
)

func init() {
	accessCmd.AddCommand(newAccessDynamicProxyCommand().cmd)
}

type accessDynamicProxyCommand struct {
	configPath string
	cmd        *cobra.Command
}

func newAccessDynamicProxyCommand() *accessDynamicProxyCommand {
	cmd := &cobra.Command{
		Use:   "dynamicProxy <configPath>",
		Short: "Launch a dynamic proxy service",
		Args:  cobra.ExactArgs(1),
	}
	command := &accessDynamicProxyCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *accessDynamicProxyCommand) run(_ *cobra.Command, args []string) {
	cmd.configPath = args[0]

	root, err := environment.LoadRoot()
	if err != nil {
		cmd.error(err)
	}

	if !root.IsEnabled() {
		tui.Error("unable to load environment; did you 'zrok enable'?", nil)
	}

	service, err := dynamicProxy.NewService(cmd.configPath)
	if err != nil {
		cmd.error(err)
	}

	dl.Infof("starting dynamicProxy service with config '%v'", cmd.configPath)

	go func() {
		if err := service.Start(); err != nil {
			cmd.error(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)
	<-c

	dl.Infof("shutting down dynamicProxy service")

	if err := service.Stop(); err != nil {
		dl.Errorf("error shutting down: %v", err)
	}
}

func (cmd *accessDynamicProxyCommand) error(err error) {
	if !panicInstead {
		tui.Error("unable to start dynamicProxy", err)
	}
	panic(err)
}
