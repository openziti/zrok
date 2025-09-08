package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gobwas/glob"
	"github.com/openziti/zrok/endpoints"
	"github.com/openziti/zrok/endpoints/drive"
	"github.com/openziti/zrok/endpoints/proxy"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/environment/env_core"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/openziti/zrok/tui"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newShare12Command().cmd)
}

type share12Command struct {
	namespace                 string
	name                      string
	backendMode               string
	insecure                  bool
	oauthProvider             string
	oauthEmailAddressPatterns []string
	oauthCheckInterval        time.Duration
	open                      bool
	accessGrants              []string
	basicAuth                 []string
	cmd                       *cobra.Command
}

func newShare12Command() *share12Command {
	cmd := &cobra.Command{
		Use:   "share12 <target>",
		Short: "Share a target resource using namespace selection",
		Args:  cobra.ExactArgs(1),
	}
	command := &share12Command{cmd: cmd}
	
	cmd.Flags().StringVar(&command.namespace, "namespace", "public", "Namespace token to use for the share")
	cmd.Flags().StringVar(&command.name, "name", "", "Reserved name to use (auto-generates if omitted)")
	cmd.Flags().StringVarP(&command.backendMode, "backend-mode", "b", "proxy", "The backend mode {proxy, web, caddy, drive}")
	cmd.Flags().BoolVar(&command.insecure, "insecure", false, "Enable insecure TLS certificate validation for <target>")
	cmd.Flags().BoolVar(&command.open, "open", false, "Enable open permission mode")
	cmd.Flags().StringArrayVar(&command.accessGrants, "access-grant", []string{}, "zrok accounts that are allowed to access this share")
	cmd.Flags().StringArrayVar(&command.basicAuth, "basic-auth", []string{}, "Basic authentication users (<username:password>,...)")
	cmd.Flags().StringVar(&command.oauthProvider, "oauth-provider", "", "Select named OAuth provider")
	cmd.Flags().StringArrayVar(&command.oauthEmailAddressPatterns, "oauth-email-address-pattern", []string{}, "Allow only email addresses matching this glob to access")
	cmd.Flags().DurationVar(&command.oauthCheckInterval, "oauth-check-interval", 3*time.Hour, "Maximum lifetime for OAuth authentication; refresh after expiry")
	cmd.MarkFlagsMutuallyExclusive("basic-auth", "oauth-provider")

	cmd.Run = command.run
	return command
}

func (cmd *share12Command) run(_ *cobra.Command, args []string) {
	root, err := environment.LoadRoot()
	if err != nil {
		cmd.error("error loading environment", err)
	}

	if !root.IsEnabled() {
		cmd.error("unable to create share", errors.New("unable to load environment; did you 'zrok enable'?"))
	}

	cmd.shareLocal(args, root)
}

func (cmd *share12Command) shareLocal(args []string, root env_core.Root) {
	var target string

	superNetwork, _ := root.SuperNetwork()

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

	case "drive":
		target = args[0]

	default:
		cmd.error("unable to create share", fmt.Errorf("invalid backend mode '%v'; expected {proxy, web, caddy, drive}", cmd.backendMode))
	}

	zif, err := root.ZitiIdentityNamed(root.EnvironmentIdentityName())
	if err != nil {
		cmd.error("unable to access ziti identity file", err)
	}

	// validate oauth email patterns
	for _, g := range cmd.oauthEmailAddressPatterns {
		_, err := glob.Compile(g)
		if err != nil {
			cmd.error(fmt.Sprintf("unable to create share, invalid oauth email glob (%v)", g), err)
		}
	}

	// create share request
	req := &sdk.Share12Request{
		EnvZId:                      root.Environment().ZitiIdentity,
		ShareMode:                   "public",
		Target:                      target,
		BackendMode:                 cmd.backendMode,
		PermissionMode:              sdk.ClosedPermissionMode,
		AccessGrants:                cmd.accessGrants,
		BasicAuthUsers:              cmd.basicAuth,
		OauthProvider:               cmd.oauthProvider,
		OauthEmailDomains:           cmd.oauthEmailAddressPatterns,
		OauthRefreshInterval:        cmd.oauthCheckInterval.String(),
		NamespaceSelections:         []*rest_model_zrok.NamespaceSelection{{NamespaceToken: cmd.namespace, Name: cmd.name}},
	}
	if cmd.open {
		req.PermissionMode = sdk.OpenPermissionMode
	}

	shr, err := sdk.CreateShare12(root, req)
	if err != nil {
		cmd.error("unable to create share", err)
	}

	logrus.Infof("created share '%v'", shr.ShareToken)
	logrus.Infof("access your zrok share at the following endpoints:")
	for _, endpoint := range shr.FrontendProxyEndpoints {
		logrus.Infof("  %v", endpoint)
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
			ShrToken:        shr.ShareToken,
			Insecure:        cmd.insecure,
			Requests:        requests,
			SuperNetwork:    superNetwork,
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
			ShrToken:     shr.ShareToken,
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
		// convert to sdk.Share for caddy backend compatibility
		sdkShare := &sdk.Share{
			Token:             shr.ShareToken,
			FrontendEndpoints: shr.FrontendProxyEndpoints,
		}
		cfg := &proxy.CaddyfileBackendConfig{
			CaddyfilePath: target,
			Shr:           sdkShare,
			Requests:      requests,
		}

		be, err := proxy.NewCaddyfileBackend(cfg)
		if err != nil {
			cmd.shutdown(root, sdkShare)
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
			ShrToken:     shr.ShareToken,
			Requests:     requests,
			SuperNetwork: superNetwork,
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
		cmd.error("invalid backend mode", nil)
	}

	// log requests
	for {
		select {
		case req := <-requests:
			logrus.Infof("%v -> %v %v", req.RemoteAddr, req.Method, req.Path)
		}
	}
}

func (cmd *share12Command) error(msg string, err error) {
	if !panicInstead {
		tui.Error(msg, err)
	}
	panic(errors.Wrap(err, msg))
}

func (cmd *share12Command) shutdown(root env_core.Root, shr interface{}) {
	switch s := shr.(type) {
	case *sdk.Share12Response:
		logrus.Debugf("shutting down share12 '%v'", s.ShareToken)
		if err := sdk.DeleteShare12(root, s.ShareToken); err != nil {
			logrus.Errorf("error shutting down share12 '%v': %v", s.ShareToken, err)
		}
	case *sdk.Share:
		logrus.Debugf("shutting down share '%v'", s.Token)
		if err := sdk.DeleteShare(root, s); err != nil {
			logrus.Errorf("error shutting down share '%v': %v", s.Token, err)
		}
	}
	logrus.Debugf("shutdown complete")
}