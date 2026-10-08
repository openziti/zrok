package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/agent/agentClient"
	"github.com/openziti/zrok/v2/agent/agentGrpc"
	"github.com/openziti/zrok/v2/cmd/zrok2/subordinate"
	"github.com/openziti/zrok/v2/endpoints"
	"github.com/openziti/zrok/v2/endpoints/drive"
	"github.com/openziti/zrok/v2/endpoints/proxy"
	"github.com/openziti/zrok/v2/endpoints/socks"
	"github.com/openziti/zrok/v2/endpoints/tcpTunnel"
	"github.com/openziti/zrok/v2/endpoints/udpTunnel"
	"github.com/openziti/zrok/v2/environment"
	"github.com/openziti/zrok/v2/environment/env_core"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
	"github.com/openziti/zrok/v2/tui"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	shareCmd.AddCommand(newSharePrivateCommand().cmd)
}

type sharePrivateCommand struct {
	backendMode  string
	shareToken   string
	headless     bool
	subordinate  bool
	forceLocal   bool
	forceAgent   bool
	insecure     bool
	open         bool
	accessGrants []string
	cmd          *cobra.Command
}

func newSharePrivateCommand() *sharePrivateCommand {
	cmd := &cobra.Command{
		Use:   "private [<target>]",
		Short: "Share a target resource privately",
		Long:  "Share a target resource privately\n\n" + exitCodesHelp,
		Args:  cobra.RangeArgs(0, 1),
	}
	command := &sharePrivateCommand{cmd: cmd}
	headless := false
	if root, err := environment.LoadRoot(); err == nil {
		headless, _ = root.Headless()
	}
	cmd.Flags().StringVarP(&command.backendMode, "backend-mode", "b", "proxy", "The backend mode {proxy, web, tcpTunnel, udpTunnel, caddy, drive, socks}")
	cmd.Flags().StringVarP(&command.shareToken, "share-token", "s", "", "Use an existing share instead of creating new")
	cmd.Flags().BoolVar(&command.headless, "headless", headless, "Disable TUI and run headless")
	cmd.Flags().BoolVar(&command.subordinate, "subordinate", false, "Enable agent mode")
	cmd.MarkFlagsMutuallyExclusive("headless", "subordinate")
	cmd.Flags().BoolVar(&command.forceLocal, "force-local", false, "Skip agent detection and force local mode")
	cmd.Flags().BoolVar(&command.forceAgent, "force-agent", false, "Skip agent detection and force agent mode")
	cmd.MarkFlagsMutuallyExclusive("force-local", "force-agent")
	cmd.Flags().BoolVar(&command.insecure, "insecure", false, "Enable insecure TLS certificate validation for <target>")
	cmd.Flags().BoolVar(&command.open, "open", false, "Enable open permission mode")
	cmd.Flags().StringArrayVar(&command.accessGrants, "access-grant", []string{}, "zrok accounts that are allowed to access this share (see --closed)")
	cmd.Run = command.run
	return command
}

func (cmd *sharePrivateCommand) run(cobraCmd *cobra.Command, args []string) {
	if cmd.subordinate {
		logrus.SetFormatter(&logrus.JSONFormatter{TimestampFormat: time.RFC3339Nano})
		dlOpts := dl.DefaultOptions().SetTrimPrefix(trimPrefix).SetLevel(slog.LevelInfo)
		dlOpts.UseJSON = true
		dl.Init(dlOpts)
	}

	root, err := environment.LoadRoot()
	if err != nil {
		cmd.error("error loading environment", err)
	}

	if !root.IsEnabled() {
		exitWithFailure("unable to load environment; did you 'zrok2 enable'?", nil)
	}

	detectAndRouteToAgent(
		cmd.subordinate, cmd.forceLocal, cmd.forceAgent,
		root,
		func() { cmd.shareLocal(cobraCmd, args, root) },
		func() { cmd.shareAgent(cobraCmd, args, root) },
	)
}

func (cmd *sharePrivateCommand) shareLocal(cobraCmd *cobra.Command, args []string, root env_core.Root) {
	var shr *sdk.Share
	var skipDelete bool
	var backendMode string

	if cmd.shareToken != "" {
		// using existing share - verify it exists and get its backend mode
		if cobraCmd.Flags().Changed("backend-mode") {
			cmd.error("unable to create share", errors.New("--backend-mode cannot be specified when using --share-token"))
		}

		shareDetail, err := sdk.GetShareDetail(root, cmd.shareToken)
		if err != nil {
			cmd.error("share not found", err)
		}
		if shareDetail.ShareMode != "private" {
			cmd.error("share is not private", errors.New("invalid share mode"))
		}

		backendMode = shareDetail.BackendMode
		shr = &sdk.Share{
			Token:             shareDetail.ShareToken,
			FrontendEndpoints: shareDetail.FrontendEndpoints,
		}
		skipDelete = true
	} else {
		backendMode = cmd.backendMode
		skipDelete = false
	}

	// validate and process backend mode (nil = allow all modes for private shares)
	target, forceHeadless, err := validateBackendMode(backendMode, args, nil)
	if err != nil {
		cmd.error("unable to create share", err)
	}
	if forceHeadless {
		cmd.headless = true
	}

	superNetwork, _ := root.SuperNetwork()

	zif, err := root.ZitiIdentityNamed(root.EnvironmentIdentityName())
	if err != nil {
		cmd.error("unable to load ziti identity configuration", err)
	}

	if shr == nil {
		// create ephemeral share (existing behavior)
		req := &sdk.ShareRequest{
			BackendMode:    sdk.BackendMode(backendMode),
			ShareMode:      sdk.PrivateShareMode,
			Target:         target,
			PermissionMode: sdk.ClosedPermissionMode,
			AccessGrants:   cmd.accessGrants,
		}
		if cmd.open {
			req.PermissionMode = sdk.OpenPermissionMode
		}

		shr, err = sdk.CreateShare(root, req)
		if err != nil {
			cmd.error("unable to create share", err)
		}
	}

	shareDescription := fmt.Sprintf("access your share with: %v", tui.Code.Render(fmt.Sprintf("zrok2 access private %v", shr.Token)))
	mdl := newShareModel(shr.Token, []string{shareDescription}, sdk.PrivateShareMode, sdk.BackendMode(backendMode))
	if !cmd.headless && !cmd.subordinate {
		proxy.SetCaddyLoggingWriter(mdl)
	}

	backend := &runningBackend{}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cmd.shutdown(root, shr, skipDelete, backend)
		os.Exit(0)
	}()

	// fail deletes a share this command created before exiting, so the next attempt creates it cleanly
	// instead of conflicting with this one
	fail := func(msg string, err error) {
		cmd.shutdown(root, shr, skipDelete, backend)
		cmd.error(msg, err)
	}

	requests := make(chan *endpoints.Request, 1024)

	var be shareBackend
	switch backendMode {
	case "proxy":
		cfg := &proxy.BackendConfig{
			IdentityPath:    zif,
			EndpointAddress: target,
			ShrToken:        shr.Token,
			Insecure:        cmd.insecure,
			Requests:        requests,
			SuperNetwork:    superNetwork,
		}

		proxyBe, err := proxy.NewBackend(cfg)
		if err != nil {
			fail("unable to create 'proxy' backend", err)
		}
		be = proxyBe

	case "web":
		cfg := &proxy.CaddyWebBackendConfig{
			IdentityPath: zif,
			WebRoot:      target,
			ShrToken:     shr.Token,
			Requests:     requests,
		}

		webBe, err := proxy.NewCaddyWebBackend(cfg)
		if err != nil {
			fail("unable to create 'web' backend", err)
		}
		be = webBe

	case "tcpTunnel":
		cfg := &tcpTunnel.BackendConfig{
			IdentityPath:    zif,
			EndpointAddress: target,
			ShrToken:        shr.Token,
			RequestsChan:    requests,
			SuperNetwork:    superNetwork,
		}

		tcpBe, err := tcpTunnel.NewBackend(cfg)
		if err != nil {
			fail("unable to create 'tcpTunnel' backend", err)
		}
		be = tcpBe

	case "udpTunnel":
		cfg := &udpTunnel.BackendConfig{
			IdentityPath:    zif,
			EndpointAddress: target,
			ShrToken:        shr.Token,
			RequestsChan:    requests,
			SuperNetwork:    superNetwork,
		}

		udpBe, err := udpTunnel.NewBackend(cfg)
		if err != nil {
			fail("unable to create 'udpTunnel' backend", err)
		}
		be = udpBe

	case "caddy":
		cfg := &proxy.CaddyfileBackendConfig{
			CaddyfilePath: target,
			Shr:           shr,
			Requests:      requests,
		}

		caddyBe, err := proxy.NewCaddyfileBackend(cfg)
		if err != nil {
			fail("unable to create 'caddy' backend", err)
		}
		be = caddyBe

	case "drive":
		cfg := &drive.BackendConfig{
			IdentityPath: zif,
			DriveRoot:    target,
			ShrToken:     shr.Token,
			Requests:     requests,
			SuperNetwork: superNetwork,
		}

		driveBe, err := drive.NewBackend(cfg)
		if err != nil {
			fail("unable to create 'drive' backend", err)
		}
		be = driveBe

	case "socks":
		cfg := &socks.BackendConfig{
			IdentityPath: zif,
			ShrToken:     shr.Token,
			Requests:     requests,
			SuperNetwork: superNetwork,
		}

		socksBe, err := socks.NewBackend(cfg)
		if err != nil {
			fail("unable to create 'socks' backend", err)
		}
		be = socksBe

	default:
		fail("unable to create share", errors.New("invalid backend mode"))
	}
	backend.set(be)

	if startsInline(backendMode) {
		if err := be.Run(); err != nil {
			fail(fmt.Sprintf("unable to start '%v' backend", backendMode), err)
		}
	} else {
		go func() {
			if err := be.Run(); err != nil && !backend.isStopping() {
				dl.Errorf("error running '%v' backend: %v", backendMode, err)
			}
		}()
	}

	if cmd.subordinate {
		data := make(map[string]interface{})
		data[subordinate.MessageKey] = subordinate.BootMessage
		data["token"] = shr.Token
		data["frontend_endpoints"] = shr.FrontendEndpoints
		jsonData, err := json.Marshal(data)
		if err != nil {
			fail("unable to create share", err)
		}
		fmt.Println(string(jsonData))
	}

	if cmd.headless && !cmd.subordinate {
		dl.Infof("allow other to access your share with the following command:\nzrok2 access private %v", shr.Token)
		for {
			select {
			case req := <-requests:
				dl.Infof("%v -> %v %v", req.RemoteAddr, req.Method, req.Path)
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
		dlOpts := dl.DefaultOptions().SetTrimPrefix(trimPrefix).SetLevel(slog.LevelInfo)
		dlOpts.CustomHandler = dl.NewPrettyHandler(slog.LevelInfo, dl.DefaultOptions().SetOutput(mdl))
		dl.Init(dlOpts)

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
			fail("An error occurred", err)
		}

		close(requests)
		cmd.shutdown(root, shr, skipDelete, backend)
	}
}

func (cmd *sharePrivateCommand) error(msg string, err error) {
	if cmd.subordinate {
		subordinateError(msg, err)
	}
	exitWithFailure(msg, err)
}

// shutdown stops the backend before deleting the share, so the ziti sdk is not left rebinding to a service
// the controller is deleting.
func (cmd *sharePrivateCommand) shutdown(root env_core.Root, shr *sdk.Share, skipDelete bool, backend *runningBackend) {
	dl.Debugf("shutting down '%v'", shr.Token)
	backend.stop()
	if !skipDelete {
		if err := sdk.DeleteShare(root, shr); err != nil {
			dl.Errorf("error shutting down '%v': %v", shr.Token, err)
		}
	}
	dl.Debugf("shutdown complete")
}

func (cmd *sharePrivateCommand) shareAgent(cobraCmd *cobra.Command, args []string, root env_core.Root) {
	var target string
	var backendMode string

	if cmd.shareToken != "" {
		// using existing share - verify it exists and get its backend mode
		if cobraCmd.Flags().Changed("backend-mode") {
			exitWithFailure("--backend-mode cannot be specified when using --share-token", nil)
		}

		shareDetail, err := sdk.GetShareDetail(root, cmd.shareToken)
		if err != nil {
			exitWithFailure("share not found", err)
		}
		if shareDetail.ShareMode != "private" {
			exitWithFailure("share is not private", nil)
		}

		backendMode = shareDetail.BackendMode
	} else {
		backendMode = cmd.backendMode
	}

	switch backendMode {
	case "proxy":
		if len(args) != 1 {
			exitWithFailure("the 'proxy' backend mode expects a <target>", nil)
		}
		v, err := parseUrl(args[0])
		if err != nil {
			exitWithFailure("invalid target endpoint URL", err)
		}
		target = v

	case "web":
		if len(args) != 1 {
			exitWithFailure("the 'web' backend mode expects a <target>", nil)
		}
		v, err := filepath.Abs(args[0])
		if err != nil {
			exitWithFailure("invalid target endpoint URL", err)
		}
		target = v

	case "tcpTunnel":
		if len(args) != 1 {
			exitWithFailure("the 'tcpTunnel' backend mode expects a <target>", nil)
		}
		target = args[0]

	case "udpTunnel":
		if len(args) != 1 {
			exitWithFailure("the 'udpTunnel' backend mode expects a <target>", nil)
		}
		target = args[0]

	case "caddy":
		if len(args) != 1 {
			exitWithFailure("the 'caddy' backend mode expects a <target>", nil)
		}
		v, err := filepath.Abs(args[0])
		if err != nil {
			exitWithFailure("invalid target endpoint URL", err)
		}
		target = v

	case "drive":
		if len(args) != 1 {
			exitWithFailure("the 'drive' backend mode expects a <target>", nil)
		}
		v, err := filepath.Abs(args[0])
		if err != nil {
			exitWithFailure("invalid target endpoint URL", err)
		}
		target = v

	case "socks":
		if len(args) != 0 {
			exitWithFailure("the 'socks' backend mode does not expect <target>", nil)
		}
		target = "socks"

	default:
		exitWithFailure(fmt.Sprintf("invalid backend mode '%v'", backendMode), nil)
	}

	client, conn, err := agentClient.NewClient(root)
	if err != nil {
		exitWithFailure("error connecting to agent", err)
	}
	defer func() { _ = conn.Close() }()

	shr, err := client.SharePrivate(context.Background(), &agentGrpc.SharePrivateRequest{
		Target:            target,
		PrivateShareToken: cmd.shareToken,
		BackendMode:       backendMode,
		Insecure:          cmd.insecure,
		Closed:            !cmd.open,
		AccessGrants:      cmd.accessGrants,
	})
	if err != nil {
		exitWithFailure("error creating share", err)
	}

	fmt.Println(shr)
}
