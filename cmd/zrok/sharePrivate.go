package main

import (
	"context"
	"encoding/json"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/openziti/zrok/agent/agentClient"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/cmd/zrok/subordinate"
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
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func init() {
	shareCmd.AddCommand(newSharePrivateCommand().cmd)
}

type sharePrivateCommand struct {
	backendMode  string
	headless     bool
	subordinate  bool
	forceLocal   bool
	forceAgent   bool
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
	headless := false
	if root, err := environment.LoadRoot(); err == nil {
		headless, _ = root.Headless()
	}
	cmd.Flags().StringVarP(&command.backendMode, "backend-mode", "b", "proxy", "The backend mode {proxy, web, tcpTunnel, udpTunnel, caddy, drive, socks, vpn}")
	cmd.Flags().BoolVar(&command.headless, "headless", headless, "Disable TUI and run headless")
	cmd.Flags().BoolVar(&command.subordinate, "subordinate", false, "Enable agent mode")
	cmd.MarkFlagsMutuallyExclusive("headless", "subordinate")
	cmd.Flags().BoolVar(&command.forceLocal, "force-local", false, "Skip agent detection and force local mode")
	cmd.Flags().BoolVar(&command.forceAgent, "force-agent", false, "Skip agent detection and force agent mode")
	cmd.MarkFlagsMutuallyExclusive("force-local", "force-agent")
	cmd.Flags().BoolVar(&command.insecure, "insecure", false, "Enable insecure TLS certificate validation for <target>")
	cmd.Flags().BoolVar(&command.closed, "closed", false, "Enable closed permission mode (see --access-grant)")
	cmd.Flags().StringArrayVar(&command.accessGrants, "access-grant", []string{}, "zrok accounts that are allowed to access this share (see --closed)")
	cmd.Run = command.run
	return command
}

func (cmd *sharePrivateCommand) run(_ *cobra.Command, args []string) {
	if cmd.subordinate {
		logrus.SetFormatter(&logrus.JSONFormatter{TimestampFormat: time.RFC3339Nano})
	}

	root, err := environment.LoadRoot()
	if err != nil {
		cmd.error("error loading environment", err)
	}

	if !root.IsEnabled() {
		tui.Error("unable to load environment; did you 'zrok enable'?", nil)
	}

	if cmd.subordinate || cmd.forceLocal {
		cmd.shareLocal(args, root)
	} else {
		agent := cmd.forceAgent
		if !cmd.forceAgent {
			agent, err = agentClient.IsAgentRunning(root)
			if err != nil {
				tui.Error("error checking if agent is running", err)
			}
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
			cmd.error("unable to create share", errors.New("the 'proxy' backend mode expects a <target>"))
		}
		v, err := parseUrl(args[0])
		if err != nil {
			cmd.error("invalid target endpoint URL", err)
		}
		target = v

	case "web":
		if len(args) != 1 {
			cmd.error("unable to create share", errors.New("the 'web' backend mode expects a <target>"))
		}
		target = args[0]

	case "tcpTunnel":
		if len(args) != 1 {
			cmd.error("unable to create share", errors.New("the 'tcpTunnel' backend mode expects a <target>"))
		}
		target = args[0]

	case "udpTunnel":
		if len(args) != 1 {
			cmd.error("unable to create share", errors.New("the 'udpTunnel' backend mode expects a <target>"))
		}
		target = args[0]

	case "caddy":
		if len(args) != 1 {
			cmd.error("unable to create share", errors.New("the 'caddy' backend mode expects a <target>"))
		}
		target = args[0]
		cmd.headless = true

	case "drive":
		if len(args) != 1 {
			cmd.error("unable to create share", errors.New("the 'drive' backend mode expects a <target>"))
		}
		target = args[0]

	case "socks":
		if len(args) != 0 {
			cmd.error("unable to create share", errors.New("the 'socks' backend mode expects a <target>"))
		}
		target = "socks"

	case "vpn":
		if len(args) == 1 {
			_, _, err := net.ParseCIDR(args[0])
			if err != nil {
				cmd.error("unable to create share", errors.New("the 'vpn' backend mode expects a valid CIDR <target>"))
			}
			target = args[0]
		} else {
			target = vpn.DefaultTarget()
		}

	default:
		cmd.error("unable to create share", fmt.Errorf("invalid backend mode '%v'; expected {proxy, web, tcpTunnel, udpTunnel, caddy, drive}", cmd.backendMode))
	}

	root, err := environment.LoadRoot()
	if err != nil {
		cmd.error("unable to load environment", err)
	}

	if !root.IsEnabled() {
		cmd.error("unable to create share", errors.New("unable to load environment; did you 'zrok enable'?"))
	}

	zif, err := root.ZitiIdentityNamed(root.EnvironmentIdentityName())
	if err != nil {
		cmd.error("unable to load ziti identity configuration", err)
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
		cmd.error("unable to create share", err)
	}

	shareDescription := fmt.Sprintf("access your share with: %v", tui.Code.Render(fmt.Sprintf("zrok access private %v", shr.Token)))
	mdl := newShareModel(shr.Token, []string{shareDescription}, sdk.PrivateShareMode, sdk.BackendMode(cmd.backendMode))
	if !cmd.headless && !cmd.subordinate {
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
			cmd.error("unable to create 'proxy' backend", err)
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
			cmd.error("unable to create 'web' backend", err)
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
			cmd.error("unable to create 'tcpTunnel' backend", err)
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
			cmd.error("unable to create 'udpTunnel' backend", err)
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
			cmd.error("unable to create 'caddy' backend", err)
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
			cmd.error("unable to create 'drive' backend", err)
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
			cmd.error("unable to create 'socks' backend", err)
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
			cmd.error("unable to create 'vpn' backend", err)
		}

		go func() {
			if err := be.Run(); err != nil {
				logrus.Errorf("error running VPN backend: %v", err)
			}
		}()

	default:
		cmd.error("unable to create share", errors.New("invalid backend mode"))
	}

	if cmd.subordinate {
		data := make(map[string]interface{})
		data[subordinate.MessageKey] = subordinate.BootMessage
		data["token"] = shr.Token
		data["frontend_endpoints"] = shr.FrontendEndpoints
		jsonData, err := json.Marshal(data)
		if err != nil {
			cmd.error("unable to create share", err)
		}
		fmt.Println(string(jsonData))
	}

	if cmd.headless && !cmd.subordinate {
		logrus.Infof("allow other to access your share with the following command:\nzrok access private %v", shr.Token)
		for {
			select {
			case req := <-requests:
				logrus.Infof("%v -> %v %v", req.RemoteAddr, req.Method, req.Path)
			}
		}

	} else if cmd.subordinate {
		for {
			select {
			case req := <-requests:
				data := make(map[string]interface{})
				data[subordinate.MessageKey] = "access"
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

func (cmd *sharePrivateCommand) error(msg string, err error) {
	if cmd.subordinate {
		subordinateError(errors.Wrap(err, msg))
	}
	if !panicInstead {
		tui.Error(msg, err)
	}
	panic(errors.Wrap(err, msg))
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
