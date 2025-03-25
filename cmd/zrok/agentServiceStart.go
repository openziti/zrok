//go:build windows

package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"os"
	"path/filepath"
	"time"
)

func init() {
	agentServiceCmd.AddCommand(newAgentServiceStartCommand().cmd)
}

type agentServiceStartCommand struct {
	cmd *cobra.Command
}

func newAgentServiceStartCommand() *agentServiceStartCommand {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the agent as a service (on Windows)",
		Args:  cobra.NoArgs,
	}
	out := &agentServiceStartCommand{cmd: cmd}
	cmd.Run = out.run
	return out
}

func (cmd *agentServiceStartCommand) run(_ *cobra.Command, _ []string) {
	elog, err := eventlog.Open(agentServiceName)
	if err != nil {
		panic(err)
	}
	defer func() { _ = elog.Close() }()

	zrokDir, err := cmd.zrokDir()
	if err == nil {
		_ = elog.Info(1, fmt.Sprintf("zrokDir is set to '%v'", zrokDir))
	} else {
		_ = elog.Error(1, fmt.Sprintf("error getting zrokDir: %v", err))
	}

	if err := cmd.logToFile(); err != nil {
		_ = elog.Error(1, fmt.Sprintf("error logging to file: %v", err))
	}

	logrus.Infof("trying to start")

	if err := svc.Run(agentServiceName, &zrokAgentSvc{zrokDir: zrokDir, elog: elog}); err != nil {
		_ = elog.Error(1, fmt.Sprintf("service start failed: %v", err))
		logrus.Errorf("service start failed: %v", err)
	}

	_ = elog.Info(1, fmt.Sprintf("zrok agent service stopped"))
}

func (cmd *agentServiceStartCommand) logToFile() error {
	zrokDir, err := cmd.zrokDir()
	if err != nil {
		return err
	}
	logFPath := filepath.Join(zrokDir, "agent.log")
	_ = os.MkdirAll(filepath.Dir(logFPath), 0755)
	logF, err := os.OpenFile(logFPath, os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	logrus.SetOutput(logF)
	return nil
}

func (cmd *agentServiceStartCommand) zrokDir() (string, error) {
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".zrok"), nil
	} else {
		return "", err
	}
}

type zrokAgentSvc struct {
	zrokDir string
	elog    *eventlog.Log
}

func (za *zrokAgentSvc) Execute(_ []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	changes <- svc.Status{State: svc.StartPending}

	//root, err := environment.LoadRoot()
	//if err != nil {
	//	_ = za.elog.Error(1, fmt.Sprintf("error loading environment: %v", err))
	//	logrus.Errorf("error loading environment: %v", err)
	//	os.Exit(1)
	//}
	//
	//if !root.IsEnabled() {
	//	_ = za.elog.Error(1, "unable to load environment; did you 'zrok enable'?")
	//	logrus.Error("unable to load environment; did you 'zrok enable'?")
	//	os.Exit(1)
	//}
	//
	//_ = za.elog.Info(1, "loaded root")
	//
	//cfg := agent.DefaultConfig()
	//a, err := agent.NewAgent(cfg, root)
	//if err != nil {
	//	_ = za.elog.Error(1, fmt.Sprintf("error creating agent: %v", err))
	//	logrus.Errorf("error creating agent: %v", err)
	//	os.Exit(1)
	//}
	//
	//_ = za.elog.Info(1, "preparing to run agent")
	//
	//time.Sleep(5 * time.Second)
	//
	//go func() {
	//	if err := a.Run(); err != nil {
	//		_ = za.elog.Error(1, fmt.Sprintf("agent aborted: %v", err))
	//		logrus.Errorf("agent aborted: %v", err)
	//		os.Exit(1)
	//	}
	//}()

	tick := time.Tick(10 * time.Second)

	changes <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}

	logrus.Info("looping")
loop:
	for {
		select {
		case <-tick:
			logrus.Infof("heartbeat at: %v", time.Now())

		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				//a.Shutdown()
				break loop
			}
		}
	}
	logrus.Infof("closing")

	changes <- svc.Status{State: svc.StopPending}
	return
}
