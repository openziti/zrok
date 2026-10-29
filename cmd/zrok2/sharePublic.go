package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/agent/agentClient"
	"github.com/openziti/zrok/v2/agent/agentGrpc"
	"github.com/openziti/zrok/v2/cmd/zrok2/subordinate"
	"github.com/openziti/zrok/v2/endpoints"
	"github.com/openziti/zrok/v2/endpoints/proxy"
	"github.com/openziti/zrok/v2/environment"
	"github.com/openziti/zrok/v2/environment/env_core"
	"github.com/openziti/zrok/v2/internal/sharepublic"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	shareCmd.AddCommand(newSharePublicCommand().cmd)
}

type sharePublicCommand struct {
	basicAuth            []string
	nameSelections       []string
	backendMode          string
	headless             bool
	subordinate          bool
	shareToken           string
	forceLocal           bool
	forceAgent           bool
	insecure             bool
	oauthProvider        string
	oauthEmailDomains    []string
	oauthRefreshInterval time.Duration
	open                 bool
	accessGrants         []string
	cmd                  *cobra.Command
}

func newSharePublicCommand() *sharePublicCommand {
	cmd := &cobra.Command{
		Use:   "public <target>",
		Short: "Share a target resource publicly",
		Long:  "Share a target resource publicly\n\n" + exitCodesHelp,
		Args:  cobra.ExactArgs(1),
	}
	command := &sharePublicCommand{cmd: cmd}
	defaultNameSelections := []string{"public"}
	if root, err := environment.LoadRoot(); err == nil {
		rootNamespace, _ := root.DefaultNamespace()
		defaultNameSelections = []string{rootNamespace}
	}
	headless := false
	if root, err := environment.LoadRoot(); err == nil {
		headless, _ = root.Headless()
	}
	cmd.Flags().StringArrayVarP(&command.nameSelections, "name-selection", "n", defaultNameSelections, "Selected frontends to use for the share")
	cmd.Flags().StringVarP(&command.backendMode, "backend-mode", "b", "proxy", "The backend mode {proxy, web, caddy, drive}")
	cmd.Flags().BoolVar(&command.headless, "headless", headless, "Disable TUI and run headless")
	cmd.Flags().BoolVar(&command.subordinate, "subordinate", false, "Enable agent mode")
	cmd.Flags().StringVar(&command.shareToken, "share-token", "", "Existing public share token (internal worker option)")
	_ = cmd.Flags().MarkHidden("share-token")
	cmd.MarkFlagsMutuallyExclusive("headless", "subordinate")
	cmd.Flags().BoolVar(&command.forceLocal, "force-local", false, "Skip agent detection and force local mode")
	cmd.Flags().BoolVar(&command.forceAgent, "force-agent", false, "Skip agent detection and force agent mode")
	cmd.MarkFlagsMutuallyExclusive("force-local", "force-agent")
	cmd.Flags().BoolVar(&command.insecure, "insecure", false, "Enable insecure TLS certificate validation for <target>")
	cmd.Flags().BoolVar(&command.open, "open", false, "Enable open permission mode")
	cmd.Flags().StringArrayVar(&command.accessGrants, "access-grant", []string{}, "zrok accounts that are allowed to access this share (see --closed)")
	cmd.Flags().StringArrayVar(&command.basicAuth, "basic-auth", []string{}, "Basic authentication users (<username:password>,...)")
	cmd.Flags().StringVar(&command.oauthProvider, "oauth-provider", "", "Select named OAuth provider (configured in selected frontend)")
	cmd.Flags().StringArrayVar(&command.oauthEmailDomains, "oauth-email-domain", []string{}, "Allow only email addresses matching this glob to access")
	cmd.Flags().DurationVar(&command.oauthRefreshInterval, "oauth-refresh-interval", 3*time.Hour, "Maximum lifetime for OAuth authentication; refresh after expiry")
	cmd.MarkFlagsMutuallyExclusive("basic-auth", "oauth-provider")

	cmd.Run = command.run
	return command
}

func (cmd *sharePublicCommand) run(_ *cobra.Command, args []string) {
	if err := cmd.validateWorkerOptions(); err != nil {
		cmd.error("invalid public worker options", err)
		return
	}
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
		cmd.error("unable to create share", errors.New("unable to load environment; did you 'zrok2 enable'?"))
	}

	detectAndRouteToAgent(
		cmd.subordinate, cmd.forceLocal, cmd.forceAgent,
		root,
		func() { cmd.shareLocal(args, root) },
		func() { cmd.shareAgent(args, root) },
	)
}

func (cmd *sharePublicCommand) validateWorkerOptions() error {
	supplied := cmd.shareToken != "" || cmd.cmd != nil && cmd.cmd.Flags().Changed("share-token")
	if supplied && !cmd.subordinate {
		return newPermanentError("public --share-token requires --subordinate")
	}
	if cmd.subordinate && strings.TrimSpace(cmd.shareToken) == "" {
		return newPermanentError("public --subordinate requires a nonempty --share-token")
	}
	return nil
}

// prepareRequest resolves paths in the invoking shell before either local execution or agent routing.
func (cmd *sharePublicCommand) prepareRequest(args []string) (*sdk.ShareRequest, error) {
	target, err := sharepublic.Target(cmd.backendMode, args)
	if err != nil {
		return nil, err
	}
	if cmd.backendMode != "proxy" {
		target, err = filepath.Abs(target)
		if err != nil {
			return nil, err
		}
	}
	in := sdk.ShareRequest{
		BackendMode: sdk.BackendMode(cmd.backendMode), Target: target,
		BasicAuth: cmd.basicAuth, AccessGrants: cmd.accessGrants,
		OauthProvider: cmd.oauthProvider, OauthEmailAddressPatterns: cmd.oauthEmailDomains,
		OauthRefreshInterval: cmd.oauthRefreshInterval, PermissionMode: sdk.ClosedPermissionMode,
	}
	if cmd.open {
		in.PermissionMode = sdk.OpenPermissionMode
	}
	for _, value := range cmd.nameSelections {
		selection, err := sdk.ParseNameSelection(value)
		if err != nil {
			return nil, err
		}
		in.NameSelections = append(in.NameSelections, selection)
	}
	return sharepublic.Prepare(in, "public")
}

func (cmd *sharePublicCommand) shareLocal(args []string, root env_core.Root) {
	var req *sdk.ShareRequest
	var err error
	if cmd.subordinate {
		// workers validate execution settings without resolving names or controller authentication.
		var target string
		target, err = sharepublic.Target(cmd.backendMode, args)
		if err == nil && cmd.backendMode != "proxy" {
			target, err = filepath.Abs(target)
		}
		req = &sdk.ShareRequest{ShareMode: sdk.PublicShareMode, BackendMode: sdk.BackendMode(cmd.backendMode), Target: target}
	} else {
		req, err = cmd.prepareRequest(args)
	}
	if err != nil {
		cmd.error("unable to create share", err)
		return
	}
	if cmd.backendMode == "caddy" {
		cmd.headless = true
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	err = cmd.publicLifecycle(root, req, sdk.CreateShare, sdk.DeleteShare, func(shr *sdk.Share) error {
		if cmd.subordinate {
			return cmd.runWorker(ctx, os.Stdin, req, shr, func(_ context.Context, requests chan *endpoints.Request) (shareBackend, error) {
				return newPublicBackend(root, req, shr, cmd.insecure, requests)
			})
		}
		return cmd.runBackend(ctx, root, req, shr)
	})
	if err != nil {
		cmd.error("unable to run share", err)
	}
}

// runBackend contains no controller resource operations. its caller owns the share lifecycle.
type publicBackendFactory func(context.Context, chan *endpoints.Request) (shareBackend, error)

func (cmd *sharePublicCommand) runBackend(ctx context.Context, root env_core.Root, req *sdk.ShareRequest, shr *sdk.Share) error {
	return cmd.serveBackend(ctx, req, shr, func(_ context.Context, requests chan *endpoints.Request) (shareBackend, error) {
		return newPublicBackend(root, req, shr, cmd.insecure, requests)
	})
}

// runWorker registers EOF and signal control before construction and has no controller resource capability.
func (cmd *sharePublicCommand) runWorker(parent context.Context, input io.ReadCloser, req *sdk.ShareRequest, shr *sdk.Share, construct publicBackendFactory) error {
	if shr == nil || strings.TrimSpace(shr.Token) == "" {
		return fmt.Errorf("public worker requires a share token")
	}
	ctx, stop, err := subordinate.WorkerContext(parent, input)
	if err != nil {
		return err
	}
	defer stop()
	worker := *cmd
	worker.subordinate = true
	return worker.serveBackend(ctx, req, shr, construct)
}

func (cmd *sharePublicCommand) serveBackend(ctx context.Context, req *sdk.ShareRequest, shr *sdk.Share, construct publicBackendFactory) error {
	if ctx.Err() != nil {
		return nil
	}
	mdl := newShareModel(shr.Token, shr.FrontendEndpoints, sdk.PublicShareMode, req.BackendMode)
	if !cmd.headless && !cmd.subordinate {
		proxy.SetCaddyLoggingWriter(mdl)
	}
	requests := make(chan *endpoints.Request, 1024)
	runner := &publicBackendRunner{}
	defer runner.stop()
	started := make(chan error, 1)
	go func() {
		started <- runner.start(ctx, string(req.BackendMode), func() (shareBackend, error) {
			return construct(ctx, requests)
		})
	}()
	// cancellation can arrive during a blocking constructor. mark the holder stopping before joining it.
	select {
	case err := <-started:
		if err != nil {
			return err
		}
	case <-ctx.Done():
		runner.stop()
		<-started
		return nil
	}
	if ctx.Err() != nil {
		return nil
	}
	if cmd.subordinate {
		data := map[string]interface{}{subordinate.MessageKey: subordinate.BootMessage, "token": shr.Token, "frontend_endpoints": shr.FrontendEndpoints}
		jsonData, err := json.Marshal(data)
		if err != nil {
			return err
		}
		fmt.Println(string(jsonData))
	}
	if cmd.headless && !cmd.subordinate {
		dl.Infof("access your zrok share at the following endpoints:\n %v", strings.Join(shr.FrontendEndpoints, "\n"))
	}
	if cmd.headless || cmd.subordinate {
		for {
			select {
			case <-ctx.Done():
				return nil
			case err := <-runner.failed:
				return err
			case req := <-requests:
				if cmd.subordinate {
					data := map[string]interface{}{subordinate.MessageKey: "access", "remote_address": req.RemoteAddr, "method": req.Method, "path": req.Path}
					jsonData, err := json.Marshal(data)
					if err != nil {
						return err
					}
					fmt.Println(string(jsonData))
				} else {
					dl.Infof("%v -> %v %v", req.RemoteAddr, req.Method, req.Path)
				}
			}
		}
	}
	logrus.SetOutput(mdl)
	dlOpts := dl.DefaultOptions().SetTrimPrefix(trimPrefix).SetLevel(slog.LevelInfo)
	dlOpts.CustomHandler = dl.NewPrettyHandler(slog.LevelInfo, dl.DefaultOptions().SetOutput(mdl))
	dl.Init(dlOpts)
	prg := tea.NewProgram(mdl, tea.WithAltScreen())
	mdl.prg = prg
	tuiCtx, stopTUI := context.WithCancel(ctx)
	defer stopTUI()
	failed := make(chan error, 1)
	forwarded := make(chan struct{})
	go func() {
		defer close(forwarded)
		for {
			select {
			case <-tuiCtx.Done():
				prg.Kill()
				return
			case err := <-runner.failed:
				failed <- err
				prg.Kill()
				return
			case req := <-requests:
				prg.Send(req)
			}
		}
	}()
	_, err := prg.Run()
	stopTUI()
	<-forwarded
	select {
	case backendErr := <-failed:
		return backendErr
	default:
	}
	if ctx.Err() != nil {
		return nil
	}
	return err
}

func (cmd *sharePublicCommand) error(msg string, err error) {
	if cmd.subordinate {
		subordinateError(msg, err)
	}
	exitWithFailure(msg, err)
}

func (cmd *sharePublicCommand) shareAgent(args []string, root env_core.Root) {
	prepared, err := cmd.prepareRequest(args)
	if err != nil {
		exitWithFailure("unable to create share", err)
		return
	}

	client, conn, err := agentClient.NewClient(root)
	if err != nil {
		exitWithFailure("error connecting to agent", err)
	}
	defer func() { _ = conn.Close() }()

	grpcReq := &agentGrpc.SharePublicRequest{
		Target:               prepared.Target,
		BasicAuth:            cmd.basicAuth,
		BackendMode:          cmd.backendMode,
		Insecure:             cmd.insecure,
		OauthProvider:        cmd.oauthProvider,
		OauthEmailDomains:    cmd.oauthEmailDomains,
		OauthRefreshInterval: cmd.oauthRefreshInterval.String(),
		Closed:               !cmd.open,
		AccessGrants:         cmd.accessGrants,
	}
	for _, nssStr := range cmd.nameSelections {
		nss, err := sdk.ParseNameSelection(nssStr)
		if err != nil {
			exitWithFailure(fmt.Sprintf("invalid namespace selection '%v'", nssStr), err)
		}
		grpcReq.NameSelections = append(grpcReq.NameSelections, &agentGrpc.NameSelection{
			NamespaceToken: nss.NamespaceToken,
			Name:           nss.Name,
		})
	}
	shr, err := client.SharePublic(context.Background(), grpcReq)
	if err != nil {
		exitWithFailure("error creating share", err)
	}

	fmt.Println(shr)
}
