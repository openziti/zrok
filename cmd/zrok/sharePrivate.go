package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/openziti/zrok/endpoints"
	"github.com/openziti/zrok/endpoints/drive"
	"github.com/openziti/zrok/endpoints/proxy"
	"github.com/openziti/zrok/endpoints/socks"
	"github.com/openziti/zrok/endpoints/tcpTunnel"
	"github.com/openziti/zrok/endpoints/udpTunnel"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/environment/env_core"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/openziti/zrok/tui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	shareCmd.AddCommand(newSharePrivateCommand().cmd)
}

type sharePrivateCommand struct {
	basicAuth   []string
	backendMode string
	headless    bool
	insecure    bool
	cmd         *cobra.Command
}

func newSharePrivateCommand() *sharePrivateCommand {
	cmd := &cobra.Command{
		Use:   "private [<target>]",
		Short: "Share a target resource privately",
		Args:  cobra.RangeArgs(0, 1),
	}
	command := &sharePrivateCommand{cmd: cmd}
	cmd.Flags().StringArrayVar(&command.basicAuth, "basic-auth", []string{}, "Basic authentication users (<username:password>,...")
	cmd.Flags().StringVarP(&command.backendMode, "backend-mode", "b", "proxy", "The backend mode {proxy, web, tcpTunnel, udpTunnel, caddy, drive, socks}")
	cmd.Flags().BoolVar(&command.headless, "headless", false, "Disable TUI and run headless")
	cmd.Flags().BoolVar(&command.insecure, "insecure", false, "Enable insecure TLS certificate validation for <target>")
	cmd.Run = command.run
	return command
}

func (cmd *sharePrivateCommand) run(_ *cobra.Command, args []string) {
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
		BasicAuth:   cmd.basicAuth,
		Target:      target,
	}
	shr, err := sdk.CreateShare(root, req)
	if err != nil {
		if !panicInstead {
			tui.Error("unable to create share", err)
		}
		panic(err)
	}

	shareDescription := fmt.Sprintf("access your share with: %v", tui.Code.Render(fmt.Sprintf("zrok access private %v", shr.Token)))
	mdl := newShareModel(shr.Token, []string{shareDescription}, sdk.PrivateShareMode, sdk.BackendMode(cmd.backendMode))
	if !cmd.headless {
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
