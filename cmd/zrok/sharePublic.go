package main

import (
	"context"
	"encoding/json"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gobwas/glob"
	"github.com/openziti/zrok/agent/agentClient"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/cmd/zrok/subordinate"
	"github.com/openziti/zrok/endpoints"
	"github.com/openziti/zrok/endpoints/drive"
	"github.com/openziti/zrok/endpoints/proxy"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/environment/env_core"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/openziti/zrok/tui"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func init() {
	shareCmd.AddCommand(newSharePublicCommand().cmd)
}

type sharePublicCommand struct {
	basicAuth                 []string
	frontendSelection         []string
	backendMode               string
	headless                  bool
	subordinate               bool
	forceLocal                bool
	forceAgent                bool
	insecure                  bool
	oauthProvider             string
	oauthEmailAddressPatterns []string
	oauthCheckInterval        time.Duration
	closed                    bool
	accessGrants              []string
	cmd                       *cobra.Command
}

func newSharePublicCommand() *sharePublicCommand {
	cmd := &cobra.Command{
		Use:   "public <target>",
		Short: "Share a target resource publicly",
		Args:  cobra.ExactArgs(1),
	}
	command := &sharePublicCommand{cmd: cmd}
	defaultFrontends := []string{"public"}
	if root, err := environment.LoadRoot(); err == nil {
		defaultFrontend, _ := root.DefaultFrontend()
		defaultFrontends = []string{defaultFrontend}
	}
	headless := false
	if root, err := environment.LoadRoot(); err == nil {
		headless, _ = root.Headless()
	}
	cmd.Flags().StringArrayVar(&command.frontendSelection, "frontend", defaultFrontends, "Selected frontends to use for the share")
	cmd.Flags().StringVarP(&command.backendMode, "backend-mode", "b", "proxy", "The backend mode {proxy, web, caddy, drive}")
	cmd.Flags().BoolVar(&command.headless, "headless", headless, "Disable TUI and run headless")
	cmd.Flags().BoolVar(&command.subordinate, "subordinate", false, "Enable agent mode")
	cmd.MarkFlagsMutuallyExclusive("headless", "subordinate")
	cmd.Flags().BoolVar(&command.forceLocal, "force-local", false, "Skip agent detection and force local mode")
	cmd.Flags().BoolVar(&command.forceAgent, "force-agent", false, "Skip agent detection and force agent mode")
	cmd.MarkFlagsMutuallyExclusive("force-local", "force-agent")
	cmd.Flags().BoolVar(&command.insecure, "insecure", false, "Enable insecure TLS certificate validation for <target>")
	cmd.Flags().BoolVar(&command.closed, "closed", false, "Enable closed permission mode (see --access-grant)")
	cmd.Flags().StringArrayVar(&command.accessGrants, "access-grant", []string{}, "zrok accounts that are allowed to access this share (see --closed)")
	cmd.Flags().StringArrayVar(&command.basicAuth, "basic-auth", []string{}, "Basic authentication users (<username:password>,...)")
	cmd.Flags().StringVar(&command.oauthProvider, "oauth-provider", "", "Enable OAuth provider [google, github]")
	cmd.Flags().StringArrayVar(&command.oauthEmailAddressPatterns, "oauth-email-address-patterns", []string{}, "Allow only these email domain globs to authenticate via OAuth")
	cmd.Flags().DurationVar(&command.oauthCheckInterval, "oauth-check-interval", 3*time.Hour, "Maximum lifetime for OAuth authentication; reauthenticate after expiry")
	cmd.MarkFlagsMutuallyExclusive("basic-auth", "oauth-provider")

	cmd.Run = command.run
	return command
}

func (cmd *sharePublicCommand) run(_ *cobra.Command, args []string) {
	if cmd.subordinate {
		logrus.SetFormatter(&logrus.JSONFormatter{TimestampFormat: time.RFC3339Nano})
	}

	root, err := environment.LoadRoot()
	if err != nil {
		cmd.error("error loading environment", err)
	}

	if !root.IsEnabled() {
		cmd.error("unable to create share", errors.New("unable to load environment; did you 'zrok enable'?"))
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

func (cmd *sharePublicCommand) shareLocal(args []string, root env_core.Root) {
	var target string

	switch cmd.backendMode {
	case "proxy":
		v, err := parseUrl(args[0])
		if err != nil {
			cmd.error("invalid target endpoint URL", err)
		}
		target = v

	case "web":
		target = args[0]

	case "caddy":
		target = args[0]
		cmd.headless = true

	case "drive":
		target = args[0]

	default:
		cmd.error("unable to create share", fmt.Errorf("invalid backend mode '%v'; expected {proxy, web, caddy, drive}", cmd.backendMode))
	}

	zif, err := root.ZitiIdentityNamed(root.EnvironmentIdentityName())
	if err != nil {
		cmd.error("unable to access ziti identity file", err)
	}

	req := &sdk.ShareRequest{
		BackendMode: sdk.BackendMode(cmd.backendMode),
		ShareMode:   sdk.PublicShareMode,
		Frontends:   cmd.frontendSelection,
		BasicAuth:   cmd.basicAuth,
		Target:      target,
	}
	if cmd.closed {
		req.PermissionMode = sdk.ClosedPermissionMode
		req.AccessGrants = cmd.accessGrants
	}
	if cmd.oauthProvider != "" {
		req.OauthProvider = cmd.oauthProvider
		req.OauthEmailAddressPatterns = cmd.oauthEmailAddressPatterns
		req.OauthAuthorizationCheckInterval = cmd.oauthCheckInterval

		for _, g := range cmd.oauthEmailAddressPatterns {
			_, err := glob.Compile(g)
			if err != nil {
				cmd.error(fmt.Sprintf("unable to create share, invalid oauth email glob (%v)", g), err)
			}
		}
	}
	shr, err := sdk.CreateShare(root, req)
	if err != nil {
		cmd.error("unable to create share", err)
	}

	mdl := newShareModel(shr.Token, shr.FrontendEndpoints, sdk.PublicShareMode, sdk.BackendMode(cmd.backendMode))
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
			cmd.error("unable to create proxy backend", err)
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
			cmd.error("unable to create web backend", err)
		}

		go func() {
			if err := be.Run(); err != nil {
				logrus.Errorf("error running http web backend: %v", err)
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
			cmd.error("unable to create caddy backend", err)
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
			cmd.error("unable to create drive backend", err)
		}

		go func() {
			if err := be.Run(); err != nil {
				logrus.Errorf("error running drive backend: %v", err)
			}
		}()

	default:
		tui.Error("invalid backend mode", nil)
	}

	if cmd.subordinate {
		data := make(map[string]interface{})
		data[subordinate.MessageKey] = subordinate.BootMessage
		data["token"] = shr.Token
		data["frontend_endpoints"] = shr.FrontendEndpoints
		jsonData, err := json.Marshal(data)
		if err != nil {
			cmd.error("unable to marshal", err)
		}
		fmt.Println(string(jsonData))
	}

	if cmd.headless && !cmd.subordinate {
		logrus.Infof("access your zrok share at the following endpoints:\n %v", strings.Join(shr.FrontendEndpoints, "\n"))
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

func (cmd *sharePublicCommand) error(msg string, err error) {
	if cmd.subordinate {
		subordinateError(errors.Wrap(err, msg))
	}
	if !panicInstead {
		tui.Error(msg, err)
	}
	panic(errors.Wrap(err, msg))
}

func (cmd *sharePublicCommand) shutdown(root env_core.Root, shr *sdk.Share) {
	logrus.Debugf("shutting down '%v'", shr.Token)
	if err := sdk.DeleteShare(root, shr); err != nil {
		logrus.Errorf("error shutting down '%v': %v", shr.Token, err)
	}
	logrus.Debugf("shutdown complete")
}

func (cmd *sharePublicCommand) shareAgent(args []string, root env_core.Root) {
	var target string

	switch cmd.backendMode {
	case "proxy":
		v, err := parseUrl(args[0])
		if err != nil {
			if !panicInstead {
				tui.Error("invalid target endpoint URL", err)
			}
			panic(err)
		}
		target = v

	case "web":
		v, err := filepath.Abs(args[0])
		if err != nil {
			if !panicInstead {
				tui.Error("invalid target endpoint URL", err)
			}
			panic(err)
		}
		target = v

	case "caddy":
		v, err := filepath.Abs(args[0])
		if err != nil {
			if !panicInstead {
				tui.Error("invalid target endpoint URL", err)
			}
			panic(err)
		}
		target = v

	case "drive":
		v, err := filepath.Abs(args[0])
		if err != nil {
			if !panicInstead {
				tui.Error("invalid target endpoint URL", err)
			}
			panic(err)
		}
		target = v

	default:
		tui.Error(fmt.Sprintf("invalid backend mode '%v'", cmd.backendMode), nil)
	}

	client, conn, err := agentClient.NewClient(root)
	if err != nil {
		tui.Error("error connecting to agent", err)
	}
	defer func() { _ = conn.Close() }()

	shr, err := client.SharePublic(context.Background(), &agentGrpc.SharePublicRequest{
		Target:                    target,
		BasicAuth:                 cmd.basicAuth,
		FrontendSelection:         cmd.frontendSelection,
		BackendMode:               cmd.backendMode,
		Insecure:                  cmd.insecure,
		OauthProvider:             cmd.oauthProvider,
		OauthEmailAddressPatterns: cmd.oauthEmailAddressPatterns,
		OauthCheckInterval:        cmd.oauthCheckInterval.String(),
		Closed:                    cmd.closed,
		AccessGrants:              cmd.accessGrants,
	})
	if err != nil {
		tui.Error("error creating share", err)
	}

	fmt.Println(shr)
}
