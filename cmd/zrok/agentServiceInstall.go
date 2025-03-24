//go:build windows

package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
	"os"
	"path/filepath"
)

func init() {
	agentServiceCmd.AddCommand(newAgentServiceInstallCommand().cmd)
}

type agentServiceInstallCommand struct {
	cmd *cobra.Command
}

func newAgentServiceInstallCommand() *agentServiceInstallCommand {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install the zrok agent as a service (on Windows)",
		Args:  cobra.NoArgs,
	}
	out := &agentServiceInstallCommand{cmd: cmd}
	cmd.Run = out.run
	return out
}

func (cmd *agentServiceInstallCommand) run(_ *cobra.Command, _ []string) {
	exePath, err := filepath.Abs(os.Args[0])
	if err != nil {
		panic(err)
	}

	svcMgr, err := mgr.Connect()
	if err != nil {
		panic(err)
	}
	defer func() { _ = svcMgr.Disconnect() }()

	svc, err := svcMgr.OpenService("zrokagent")
	if err == nil {
		_ = svc.Close()
		logrus.Infof("service already exists!")
		os.Exit(1)
	}

	svcCfg := mgr.Config{DisplayName: "zrok Agent"}
	svc, err = svcMgr.CreateService("zrokagent", exePath, svcCfg, "agent", "service", "start")
	if err != nil {
		panic(err)
	}
	defer func() { _ = svc.Close() }()

	err = eventlog.InstallAsEventCreate("zrokagent", eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		_ = svc.Delete()
		panic(err)
	}

	logrus.Infof("zrok agent service installled")
}
