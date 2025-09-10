package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/openziti/zrok/endpoints/dynamicProxy"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/tui"
	"github.com/sirupsen/logrus"
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

	logrus.Infof("starting dynamic proxy service with config: %v", cmd.configPath)

	go func() {
		if err := service.Start(); err != nil {
			cmd.error(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)
	<-c

	logrus.Infof("shutting down dynamic proxy service")

	if err := service.Stop(); err != nil {
		logrus.Errorf("error shutting down: %v", err)
	}
}

func (cmd *accessDynamicProxyCommand) error(err error) {
	if !panicInstead {
		tui.Error("unable to start dynamic proxy", err)
	}
	panic(err)
}
