package main

import (
	"github.com/openziti/zrok/agent"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/tui"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	agentCmd.AddCommand(newAgentStartCommand().cmd)
}

type agentStartCommand struct {
	cmd              *cobra.Command
	consoleAddress   string
	consoleStartPort uint16
	consoleEndPort   uint16
}

func newAgentStartCommand() *agentStartCommand {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a zrok agent",
		Args:  cobra.NoArgs,
	}
	command := &agentStartCommand{cmd: cmd}
	cmd.Run = command.run
	cmd.Flags().StringVar(&command.consoleAddress, "console-address", "127.0.0.1", "gRPC gateway address")
	cmd.Flags().Uint16Var(&command.consoleStartPort, "console-start-port", 8888, "gRPC gateway starting port")
	cmd.Flags().Uint16Var(&command.consoleEndPort, "console-end-port", 8988, "gRPC gateway ending port")
	return command
}

func (cmd *agentStartCommand) run(_ *cobra.Command, _ []string) {
	root, err := environment.LoadRoot()
	if err != nil {
		tui.Error("error loading zrokdir", err)
	}

	if !root.IsEnabled() {
		tui.Error("unable to load environment; did you 'zrok enable'?", nil)
	}

	cfg := agent.DefaultConfig()
	cfg.ConsoleAddress = cmd.consoleAddress
	cfg.ConsoleStartPort = cmd.consoleStartPort
	cfg.ConsoleEndPort = cmd.consoleEndPort
	a, err := agent.NewAgent(cfg, root)
	if err != nil {
		tui.Error("error creating agent", err)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cmd.shutdown(a)
		os.Exit(0)
	}()

	if err := a.Run(); err != nil {
		tui.Error("agent aborted", err)
	}
}

func (cmd *agentStartCommand) shutdown(a *agent.Agent) {
	a.Shutdown()
}
