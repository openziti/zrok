package main

import (
	"context"
	"fmt"
	"github.com/openziti/zrok/agent/agentClient"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/endpoints/vpn"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/tui"
	"github.com/spf13/cobra"
	"net"
	"path/filepath"
)

func init() {
	agentShareCmd.AddCommand(newAgentSharePrivateCommand().cmd)
}

type agentSharePrivateCommand struct {
	backendMode  string
	insecure     bool
	closed       bool
	accessGrants []string
	cmd          *cobra.Command
}

func newAgentSharePrivateCommand() *agentSharePrivateCommand {
	cmd := &cobra.Command{
		Use:   "private <target>",
		Short: "Create a private share in the zrok Agent",
		Args:  cobra.RangeArgs(0, 1),
	}
	command := &agentSharePrivateCommand{cmd: cmd}
	cmd.Flags().StringVarP(&command.backendMode, "backend-mode", "b", "proxy", "The backend mode {proxy, web, tcpTunnel, udpTunnel, caddy, drive, socks, vpn}")
	cmd.Flags().BoolVar(&command.insecure, "insecure", false, "Enable insecure TLS certificate validation for <target>")
	cmd.Flags().BoolVar(&command.closed, "closed", false, "Enable closed permission mode (see --access-grant)")
	cmd.Flags().StringArrayVar(&command.accessGrants, "access-grant", []string{}, "zrok accounts that are allowed to access this share (see --closed)")
	cmd.Run = command.run
	return command
}

func (cmd *agentSharePrivateCommand) run(_ *cobra.Command, args []string) {
	var target string

	switch cmd.backendMode {
	case "proxy":
		if len(args) != 1 {
			tui.Error("the 'proxy' backend mode expects a <target>", nil)
		}
		v, err := parseUrl(args[0])
		if err != nil {
			if !panicInstead {
				tui.Error("invalid target endpoint URL", err)
			}
			panic(err)
		}
		target = v

	case "web":
		if len(args) != 1 {
			tui.Error("the 'web' backend mode expects a <target>", nil)
		}
		v, err := filepath.Abs(args[0])
		if err != nil {
			if !panicInstead {
				tui.Error("invalid target endpoint URL", err)
			}
			panic(err)
		}
		target = v

	case "tcpTunnel":
		if len(args) != 1 {
			tui.Error("the 'tcpTunnel' backend mode expects a <target>", nil)
		}
		target = args[0]

	case "udpTunnel":
		if len(args) != 1 {
			tui.Error("the 'udpTunnel' backend mode expects a <target>", nil)
		}
		target = args[0]

	case "caddy":
		if len(args) != 1 {
			tui.Error("the 'caddy' backend mode expects a <target>", nil)
		}
		v, err := filepath.Abs(args[0])
		if err != nil {
			if !panicInstead {
				tui.Error("invalid target endpoint URL", err)
			}
			panic(err)
		}
		target = v

	case "drive":
		if len(args) != 1 {
			tui.Error("the 'drive' backend mode expects a <target>", nil)
		}
		v, err := filepath.Abs(args[0])
		if err != nil {
			if !panicInstead {
				tui.Error("invalid target endpoint URL", err)
			}
			panic(err)
		}
		target = v

	case "socks":
		if len(args) != 0 {
			tui.Error("the 'socks' backend mode does not expect <target>", nil)
		}
		target = "socks"

	case "vpn":
		if len(args) == 1 {
			_, _, err := net.ParseCIDR(args[0])
			if err != nil {
				tui.Error("the 'vpn' backend expect valid CIDR <target>", err)
			}
			target = args[0]
		} else {
			target = vpn.DefaultTarget()
		}

	default:
		tui.Error(fmt.Sprintf("invalid backend mode '%v'", cmd.backendMode), nil)
	}

	root, err := environment.LoadRoot()
	if err != nil {
		if !panicInstead {
			tui.Error("unable to load environment", err)
		}
		panic(err)
	}

	if !root.IsEnabled() {
		tui.Error("unable to load environment; did you 'zrok enable'?", nil)
	}

	client, conn, err := agentClient.NewClient(root)
	if err != nil {
		tui.Error("error connecting to agent", err)
	}
	defer conn.Close()

	shr, err := client.PrivateShare(context.Background(), &agentGrpc.PrivateShareRequest{
		Target:       target,
		BackendMode:  cmd.backendMode,
		Insecure:     cmd.insecure,
		Closed:       cmd.closed,
		AccessGrants: cmd.accessGrants,
	})
	if err != nil {
		tui.Error("error creating share", err)
	}

	fmt.Println(shr)
}
