package main

import (
	"context"
	"fmt"
	"github.com/openziti/zrok/agent/agentClient"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/tui"
	"github.com/spf13/cobra"
	"time"
)

func init() {
	agentShareCmd.AddCommand(newAgentSharePublicCommand().cmd)
}

type agentSharePublicCommand struct {
	basicAuth                 []string
	frontendSelection         []string
	backendMode               string
	headless                  bool
	insecure                  bool
	oauthProvider             string
	oauthEmailAddressPatterns []string
	oauthCheckInterval        time.Duration
	closed                    bool
	accessGrants              []string
	cmd                       *cobra.Command
}

func newAgentSharePublicCommand() *agentSharePublicCommand {
	cmd := &cobra.Command{
		Use:   "public <target>",
		Short: "Create a public share in the zrok Agent",
		Args:  cobra.ExactArgs(1),
	}
	command := &agentSharePublicCommand{cmd: cmd}
	defaultFrontends := []string{"public"}
	if root, err := environment.LoadRoot(); err == nil {
		defaultFrontend, _ := root.DefaultFrontend()
		defaultFrontends = []string{defaultFrontend}
	}
	cmd.Flags().StringArrayVar(&command.frontendSelection, "frontend", defaultFrontends, "Selected frontends to use for the share")
	cmd.Flags().StringVarP(&command.backendMode, "backend-mode", "b", "proxy", "The backend mode {proxy, web, caddy, drive}")
	cmd.Flags().BoolVar(&command.headless, "headless", false, "Disable TUI and run headless")
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

func (cmd *agentSharePublicCommand) run(_ *cobra.Command, args []string) {
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

	shr, err := client.PublicShare(context.Background(), &agentGrpc.PublicShareRequest{
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

	fmt.Println(shr.GetToken())
}
