package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/environment/env_core"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/openziti/zrok/tui"
	"github.com/spf13/cobra"
)

func init() {
	accessCmd.AddCommand(newAccessStdioCommand().cmd)
}

type accessStdioCommand struct {
	cmd *cobra.Command
}

func newAccessStdioCommand() *accessStdioCommand {
	cmd := &cobra.Command{
		Use:   "stdio <shareToken>",
		Short: "Access a share using stdin/stdout",
		Args:  cobra.ExactArgs(1),
	}
	command := &accessStdioCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *accessStdioCommand) run(_ *cobra.Command, args []string) {
	shrToken := args[0]

	root, err := environment.LoadRoot()
	if err != nil {
		cmd.error(err)
	}

	if !root.IsEnabled() {
		cmd.error(fmt.Errorf("unable to load environment; did you 'zrok enable'?"))
	}

	acc, err := sdk.CreateAccess(root, &sdk.AccessRequest{ShareToken: shrToken})
	if err != nil {
		cmd.error(err)
	}

	// validate backend mode is compatible with stdio
	switch acc.BackendMode {
	case sdk.UdpTunnelBackendMode:
		cmd.deleteAccess(root, acc)
		cmd.error(fmt.Errorf("'udpTunnel' backend mode is not compatible with stdio access"))
	case sdk.VpnBackendMode:
		cmd.deleteAccess(root, acc)
		cmd.error(fmt.Errorf("'vpn' backend mode is not compatible with stdio access"))
	}

	conn, err := sdk.NewDialer(shrToken, root)
	if err != nil {
		cmd.deleteAccess(root, acc)
		cmd.error(err)
	}

	// signal handling for clean shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		<-c
		_ = conn.Close()
		cmd.deleteAccess(root, acc)
		os.Exit(0)
	}()

	// bidirectional copy
	errc := make(chan error, 2)

	// stdin -> connection
	go func() {
		_, err := io.Copy(conn, os.Stdin)
		_ = conn.Close()
		errc <- err
	}()

	// connection -> stdout
	go func() {
		_, err := io.Copy(os.Stdout, conn)
		errc <- err
	}()

	// wait for either direction to complete
	<-errc

	cmd.deleteAccess(root, acc)
}

func (cmd *accessStdioCommand) deleteAccess(root env_core.Root, acc *sdk.Access) {
	if err := sdk.DeleteAccess(root, acc); err != nil {
		tui.Warning(fmt.Sprintf("error deleting access: %v", err))
	}
}

func (cmd *accessStdioCommand) error(err error) {
	if !panicInstead {
		tui.Error("unable to create stdio access", err)
	}
	panic(err)
}
