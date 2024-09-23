package main

import (
	"context"
	"encoding/json"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/openziti/zrok/agent/agentClient"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/endpoints"
	"github.com/openziti/zrok/endpoints/drive"
	"github.com/openziti/zrok/endpoints/proxy"
	"github.com/openziti/zrok/endpoints/socks"
	"github.com/openziti/zrok/endpoints/tcpTunnel"
	"github.com/openziti/zrok/endpoints/udpTunnel"
	"github.com/openziti/zrok/endpoints/vpn"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/environment/env_core"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/openziti/zrok/tui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func init() {
	shareCmd.AddCommand(newSharePrivateCommand().cmd)
}

type sharePrivateCommand struct {
	backendMode  string
	headless     bool
	agent        bool
	insecure     bool
	closed       bool
	accessGrants []string
	cmd          *cobra.Command
}

func newSharePrivateCommand() *sharePrivateCommand {
	cmd := &cobra.Command{
		Use:   "private [<target>]",
		Short: "Share a target resource privately",
		Args:  cobra.RangeArgs(0, 1),
	}
	command := &sharePrivateCommand{cmd: cmd}
	cmd.Flags().StringVarP(&command.backendMode, "backend-mode", "b", "proxy", "The backend mode {proxy, web, tcpTunnel, udpTunnel, caddy, drive, socks, vpn}")
	cmd.Flags().BoolVar(&command.headless, "headless", false, "Disable TUI and run headless")
	cmd.Flags().BoolVar(&command.agent, "agent", false, "Enable agent mode")
	cmd.MarkFlagsMutuallyExclusive("headless", "agent")
	cmd.Flags().BoolVar(&command.insecure, "insecure", false, "Enable insecure TLS certificate validation for <target>")
	cmd.Flags().BoolVar(&command.closed, "closed", false, "Enable closed permission mode (see --access-grant)")
	cmd.Flags().StringArrayVar(&command.accessGrants, "access-grant", []string{}, "zrok accounts that are allowed to access this share (see --closed)")
	cmd.Run = command.run
	return command
}

func (cmd *sharePrivateCommand) run(_ *cobra.Command, args []string) {
	root, err := environment.LoadRoot()
	if err != nil {
		if !panicInstead {
			tui.Error("error loading environment", err)
		}
		panic(err)
	}

	if !root.IsEnabled() {
		tui.Error("unable to load environment; did you 'zrok enable'?", nil)
	}

	if cmd.agent {
		cmd.shareLocal(args, root)
	} else {
		agent, err := agentClient.IsAgentRunning(root)
		if err != nil {
			tui.Error("error checking if agent is running", err)
		}
		if agent {
			cmd.shareAgent(args, root)
		} else {
			cmd.shareLocal(args, root)
		}
	}
}

func (cmd *sharePrivateCommand) shareLocal(args []string, root env_core.Root) {
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
		target = args[0]

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
		target = args[0]
		cmd.headless = true

	case "drive":
		if len(args) != 1 {
			tui.Error("the 'drive' backend mode expects a <target>", nil)
		}
		target = args[0]

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
		tui.Error(fmt.Sprintf("invalid backend mode '%v'; expected {proxy, web, tcpTunnel, udpTunnel, caddy, drive}", cmd.backendMode), nil)
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

	zif, err := root.ZitiIdentityNamed(root.EnvironmentIdentityName())
	if err != nil {
		if !panicInstead {
			tui.Error("unable to load ziti identity configuration", err)
		}
		panic(err)
	}

	req := &sdk.ShareRequest{
		BackendMode: sdk.BackendMode(cmd.backendMode),
		ShareMode:   sdk.PrivateShareMode,
		Target:      target,
	}
	if cmd.closed {
		req.PermissionMode = sdk.ClosedPermissionMode
		req.AccessGrants = cmd.accessGrants
	}
	shr, err := sdk.CreateShare(root, req)
	if err != nil {
		if !panicInstead {
			tui.Error("unable to create share", err)
		}
		panic(err)
	}

	if cmd.agent {
		data := make(map[string]interface{})
		data["token"] = shr.Token
		data["frontend_endpoints"] = shr.FrontendEndpoints
		jsonData, err := json.Marshal(data)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(jsonData))
	}

	shareDescription := fmt.Sprintf("access your share with: %v", tui.Code.Render(fmt.Sprintf("zrok access private %v", shr.Token)))
	mdl := newShareModel(shr.Token, []string{shareDescription}, sdk.PrivateShareMode, sdk.BackendMode(cmd.backendMode))
	if !cmd.headless && !cmd.agent {
		proxy.SetCaddyLoggingWriter(mdl)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cmd.shutdown(root, shr)
		os.Exit(0)
	}()

	requests := make(chan *endpoints.Request, 1024)

	switch cmd.backendMode {
	case "proxy":
		cfg := &proxy.BackendConfig{
			IdentityPath:    zif,
			EndpointAddress: target,
			ShrToken:        shr.Token,
			Insecure:        cmd.insecure,
			Requests:        requests,
		}

		be, err := proxy.NewBackend(cfg)
		if err != nil {
			if !panicInstead {
				tui.Error("error creating proxy backend", err)
			}
			panic(err)
		}

		go func() {
			if err := be.Run(); err != nil {
				logrus.Errorf("error running http proxy backend: %v", err)
			}
		}()

	case "web":
		cfg := &proxy.CaddyWebBackendConfig{
			IdentityPath: zif,
			WebRoot:      target,
			ShrToken:     shr.Token,
			Requests:     requests,
		}

		be, err := proxy.NewCaddyWebBackend(cfg)
		if err != nil {
			if !panicInstead {
				tui.Error("error creating web backend", err)
			}
			panic(err)
		}

		go func() {
			if err := be.Run(); err != nil {
				logrus.Errorf("error running http web backend: %v", err)
			}
		}()

	case "tcpTunnel":
		cfg := &tcpTunnel.BackendConfig{
			IdentityPath:    zif,
			EndpointAddress: target,
			ShrToken:        shr.Token,
			RequestsChan:    requests,
		}

		be, err := tcpTunnel.NewBackend(cfg)
		if err != nil {
			if !panicInstead {
				tui.Error("error creating tcpTunnel backend", err)
			}
			panic(err)
		}

		go func() {
			if err := be.Run(); err != nil {
				logrus.Errorf("error running tcpTunnel backend: %v", err)
			}
		}()

	case "udpTunnel":
		cfg := &udpTunnel.BackendConfig{
			IdentityPath:    zif,
			EndpointAddress: target,
			ShrToken:        shr.Token,
			RequestsChan:    requests,
		}

		be, err := udpTunnel.NewBackend(cfg)
		if err != nil {
			if !panicInstead {
				tui.Error("error creating udpTunnel backend", err)
			}
			panic(err)
		}

		go func() {
			if err := be.Run(); err != nil {
				logrus.Errorf("error running udpTunnel backend: %v", err)
			}
		}()

	case "caddy":
		cfg := &proxy.CaddyfileBackendConfig{
			CaddyfilePath: target,
			Shr:           shr,
			Requests:      requests,
		}

		be, err := proxy.NewCaddyfileBackend(cfg)
		if err != nil {
			cmd.shutdown(root, shr)
			if !panicInstead {
				tui.Error("error creating caddy backend", err)
			}
			panic(err)
		}

		go func() {
			if err := be.Run(); err != nil {
				logrus.Errorf("error running caddy backend: %v", err)
			}
		}()

	case "drive":
		cfg := &drive.BackendConfig{
			IdentityPath: zif,
			DriveRoot:    target,
			ShrToken:     shr.Token,
			Requests:     requests,
		}

		be, err := drive.NewBackend(cfg)
		if err != nil {
			if !panicInstead {
				tui.Error("error creating drive backend", err)
			}
			panic(err)
		}

		go func() {
			if err := be.Run(); err != nil {
				logrus.Errorf("error running drive backend: %v", err)
			}
		}()

	case "socks":
		cfg := &socks.BackendConfig{
			IdentityPath: zif,
			ShrToken:     shr.Token,
			Requests:     requests,
		}

		be, err := socks.NewBackend(cfg)
		if err != nil {
			if !panicInstead {
				tui.Error("error creating socks backend", err)
			}
			panic(err)
		}

		go func() {
			if err := be.Run(); err != nil {
				logrus.Errorf("error running socks backend: %v", err)
			}
		}()

	case "vpn":
		cfg := &vpn.BackendConfig{
			IdentityPath:    zif,
			EndpointAddress: target,
			ShrToken:        shr.Token,
			RequestsChan:    requests,
		}

		be, err := vpn.NewBackend(cfg)
		if err != nil {
			if !panicInstead {
				tui.Error("error creating VPN backend", err)
			}
			panic(err)
		}

		go func() {
			if err := be.Run(); err != nil {
				logrus.Errorf("error running VPN backend: %v", err)
			}
		}()

	default:
		tui.Error("invalid backend mode", nil)
	}

	if cmd.headless {
		logrus.Infof("allow other to access your share with the following command:\nzrok access private %v", shr.Token)
		for {
			select {
			case req := <-requests:
				logrus.Infof("%v -> %v %v", req.RemoteAddr, req.Method, req.Path)
			}
		}

	} else if cmd.agent {
		for {
			select {
			case req := <-requests:
				data := make(map[string]interface{})
				data["remote_address"] = req.RemoteAddr
				data["method"] = req.Method
				data["path"] = req.Path
				jsonData, err := json.Marshal(data)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(string(jsonData))
			}
		}

	} else {
		logrus.SetOutput(mdl)
		prg := tea.NewProgram(mdl, tea.WithAltScreen())
		mdl.prg = prg

		go func() {
			for {
				select {
				case req := <-requests:
					prg.Send(req)
				}
			}
		}()

		if _, err := prg.Run(); err != nil {
			tui.Error("An error occurred", err)
		}

		close(requests)
		cmd.shutdown(root, shr)
	}
}

func (cmd *sharePrivateCommand) shutdown(root env_core.Root, shr *sdk.Share) {
	logrus.Debugf("shutting down '%v'", shr.Token)
	if err := sdk.DeleteShare(root, shr); err != nil {
		logrus.Errorf("error shutting down '%v': %v", shr.Token, err)
	}
	logrus.Debugf("shutdown complete")
}

func (cmd *sharePrivateCommand) shareAgent(args []string, root env_core.Root) {
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

	client, conn, err := agentClient.NewClient(root)
	if err != nil {
		tui.Error("error connecting to agent", err)
	}
	defer func() { _ = conn.Close() }()

	shr, err := client.SharePrivate(context.Background(), &agentGrpc.SharePrivateRequest{
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
