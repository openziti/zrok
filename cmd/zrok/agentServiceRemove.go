//go:build windows

package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

func init() {
	agentServiceCmd.AddCommand(newAgentServiceRemoveCommand().cmd)
}

type agentServiceRemoveCommand struct {
	cmd *cobra.Command
}

func newAgentServiceRemoveCommand() *agentServiceRemoveCommand {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove the zrok agent service (on Windows)",
		Args:  cobra.NoArgs,
	}
	out := &agentServiceRemoveCommand{cmd: cmd}
	cmd.Run = out.run
	return out
}

func (cmd *agentServiceRemoveCommand) run(_ *cobra.Command, _ []string) {
	svcMgr, err := mgr.Connect()
	if err != nil {
		panic(err)
	}
	defer func() { _ = svcMgr.Disconnect() }()

	svc, err := svcMgr.OpenService(agentServiceName)
	if err != nil {
		panic(err)
	}
	defer func() { _ = svc.Close() }()

	if err := svc.Delete(); err == nil {
		logrus.Infof("deleted zrok agent service")
	} else {
		logrus.Errorf("error deleting zrok agent service: %v", err)
	}

	if err := eventlog.Remove(agentServiceName); err == nil {
		logrus.Infof("removed zrok agent event log")
	} else {
		logrus.Errorf("error removing zrok agent event log: %v", err)
	}
}
