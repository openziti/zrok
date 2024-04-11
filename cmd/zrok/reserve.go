package main

import (
	"encoding/json"
	"fmt"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/openziti/zrok/tui"
	"github.com/openziti/zrok/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"slices"
	"time"
)

func init() {
	rootCmd.AddCommand(newReserveCommand().cmd)
}

type reserveCommand struct {
	uniqueName                string
	basicAuth                 []string
	frontendSelection         []string
	backendMode               string
	jsonOutput                bool
	oauthProvider             string
	oauthEmailAddressPatterns []string
	oauthCheckInterval        time.Duration
	closed                    bool
	accessGrants              []string
	cmd                       *cobra.Command
}

func newReserveCommand() *reserveCommand {
	cmd := &cobra.Command{
		Use:   "reserve <public|private> [<target>]",
		Short: "Create a reserved share",
		Args:  cobra.RangeArgs(1, 2),
	}
	command := &reserveCommand{cmd: cmd}
	cmd.Flags().StringVarP(&command.uniqueName, "unique-name", "n", "", "A unique name for the reserved share (defaults to generated identifier)")
	cmd.Flags().StringArrayVar(&command.frontendSelection, "frontends", []string{"public"}, "Selected frontends to use for the share")
	cmd.Flags().StringVarP(&command.backendMode, "backend-mode", "b", "proxy", "The backend mode (public|private: proxy, web, caddy, drive) (private: tcpTunnel, udpTunnel, socks, vpn)")
	cmd.Flags().BoolVarP(&command.jsonOutput, "json-output", "j", false, "Emit JSON describing the created reserved share")
	cmd.Flags().StringArrayVar(&command.basicAuth, "basic-auth", []string{}, "Basic authentication users (<username:password>,...)")
	cmd.Flags().StringVar(&command.oauthProvider, "oauth-provider", "", "Enable OAuth provider [google, github]")
	cmd.Flags().StringArrayVar(&command.oauthEmailAddressPatterns, "oauth-email-address-patterns", []string{}, "Allow only these email domains to authenticate via OAuth")
	cmd.Flags().DurationVar(&command.oauthCheckInterval, "oauth-check-interval", 3*time.Hour, "Maximum lifetime for OAuth authentication; reauthenticate after expiry")
	cmd.MarkFlagsMutuallyExclusive("basic-auth", "oauth-provider")
	cmd.Flags().BoolVar(&command.closed, "closed", false, "Enable closed permission mode (see --access-grant)")
	cmd.Flags().StringArrayVar(&command.accessGrants, "access-grant", []string{}, "zrok accounts that are allowed to access this share (see --closed)")

	cmd.Run = command.run
	return command
}

func (cmd *reserveCommand) run(_ *cobra.Command, args []string) {
	shareMode := sdk.ShareMode(args[0])
	privateOnlyModes := []string{"tcpTunnel", "udpTunnel", "socks", "vpn"}
	if shareMode != sdk.PublicShareMode && shareMode != sdk.PrivateShareMode {
		tui.Error("invalid sharing mode; expecting 'public' or 'private'", nil)
	} else if shareMode == sdk.PublicShareMode && slices.Contains(privateOnlyModes, cmd.backendMode) {
		tui.Error(fmt.Sprintf("invalid sharing mode for a %s share: %s", sdk.PublicShareMode, cmd.backendMode), nil)
	}

	if cmd.uniqueName != "" && !util.IsValidUniqueName(cmd.uniqueName) {
		tui.Error("invalid unique name; must be lowercase alphanumeric, between 4 and 32 characters in length, screened for profanity", nil)
	}

	var target string
	switch cmd.backendMode {
	case "proxy":
		if len(args) != 2 {
			tui.Error("the 'proxy' backend mode expects a <target>", nil)
		}
		v, err := parseUrl(args[1])
		if err != nil {
			tui.Error("invalid target endpoint URL", err)
		}
		target = v

	case "web":
		if len(args) != 2 {
			tui.Error("the 'web' backend mode expects a <target>", nil)
		}
		target = args[1]

	case "tcpTunnel":
		if len(args) != 2 {
			tui.Error("the 'tcpTunnel' backend mode expects a <target>", nil)
		}
		target = args[1]

	case "udpTunnel":
		if len(args) != 2 {
			tui.Error("the 'udpTunnel' backend mode expects a <target>", nil)
		}
		target = args[1]

	case "caddy":
		if len(args) != 2 {
			tui.Error("the 'caddy' backend mode expects a <target>", nil)
		}
		target = args[1]

	case "drive":
		if len(args) != 2 {
			tui.Error("the 'drive' backend mode expects a <target>", nil)
		}
		target = args[1]

	case "socks":
		if len(args) != 1 {
			tui.Error("the 'socks' backend mode does not expect <target>", nil)
		}

	case "vpn":
		target = "vpn"

	default:
		tui.Error(fmt.Sprintf("invalid backend mode '%v'; "+
			"expected {proxy, web, tcpTunnel, udpTunnel, caddy, drive, socks, vpn}", cmd.backendMode), nil)
	}

	env, err := environment.LoadRoot()
	if err != nil {
		tui.Error("error loading environment", err)
	}

	if !env.IsEnabled() {
		tui.Error("unable to load environment; did you 'zrok enable'?", nil)
	}

	req := &sdk.ShareRequest{
		Reserved:    true,
		UniqueName:  cmd.uniqueName,
		BackendMode: sdk.BackendMode(cmd.backendMode),
		ShareMode:   shareMode,
		BasicAuth:   cmd.basicAuth,
		Target:      target,
	}
	if shareMode == sdk.PublicShareMode {
		req.Frontends = cmd.frontendSelection
	}
	if cmd.oauthProvider != "" {
		if shareMode != sdk.PublicShareMode {
			tui.Error("--oauth-provider only supported for public shares", nil)
		}
		req.OauthProvider = cmd.oauthProvider
		req.OauthEmailAddressPatterns = cmd.oauthEmailAddressPatterns
		req.OauthAuthorizationCheckInterval = cmd.oauthCheckInterval
	}
	if cmd.closed {
		req.PermissionMode = sdk.ClosedPermissionMode
		req.AccessGrants = cmd.accessGrants
	}
	shr, err := sdk.CreateShare(env, req)
	if err != nil {
		tui.Error("unable to create share", err)
	}

	if !cmd.jsonOutput {
		logrus.Infof("your reserved share token is '%v'", shr.Token)
		for _, fpe := range shr.FrontendEndpoints {
			logrus.Infof("reserved frontend endpoint: %v", fpe)
		}
	} else {
		out, err := json.Marshal(shr)
		if err != nil {
			tui.Error("error emitting JSON", err)
		}
		fmt.Println(string(out))
	}
}
