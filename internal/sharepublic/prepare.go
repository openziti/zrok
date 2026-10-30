// package sharepublic prepares public requests without controller or environment access.
package sharepublic

import (
	"fmt"
	"math"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/gobwas/glob"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
)

// Prepare returns an operation-owned request. callers resolve file paths and the default namespace.
func Prepare(in sdk.ShareRequest, defaultNamespace string) (*sdk.ShareRequest, error) {
	target, err := Target(string(in.BackendMode), []string{in.Target})
	if err != nil {
		return nil, err
	}
	if len(in.BasicAuth) > 0 && in.OauthProvider != "" {
		return nil, fmt.Errorf("basic authentication and oauth provider are mutually exclusive")
	}
	for _, user := range in.BasicAuth {
		if len(strings.Split(user, ":")) != 2 {
			return nil, fmt.Errorf("invalid username:password '%v'", user)
		}
	}
	if in.OauthProvider != "" {
		for _, pattern := range in.OauthEmailAddressPatterns {
			if _, err := glob.Compile(pattern); err != nil {
				return nil, fmt.Errorf("invalid oauth email glob '%v': %w", pattern, err)
			}
		}
	}
	out := in
	out.ShareMode = sdk.PublicShareMode
	out.Target = target
	out.NameSelections = slices.Clone(in.NameSelections)
	out.BasicAuth = slices.Clone(in.BasicAuth)
	out.AccessGrants = slices.Clone(in.AccessGrants)
	out.OauthEmailAddressPatterns = slices.Clone(in.OauthEmailAddressPatterns)
	if len(out.NameSelections) == 0 {
		selection, err := sdk.ParseNameSelection(defaultNamespace)
		if err != nil {
			return nil, fmt.Errorf("invalid default name selection '%v': %w", defaultNamespace, err)
		}
		out.NameSelections = []sdk.NameSelection{selection}
	}
	for _, selection := range out.NameSelections {
		if strings.TrimSpace(selection.NamespaceToken) == "" {
			return nil, fmt.Errorf("invalid name selection: namespace must not be empty")
		}
	}
	if out.OauthProvider == "" {
		out.OauthEmailAddressPatterns = nil
		out.OauthRefreshInterval = 0
	}
	return &out, nil
}

// Target preserves the public backend set and port/URL interpretation. file paths are resolved by adapters.
func Target(mode string, args []string) (string, error) {
	switch mode {
	case "proxy", "web", "caddy", "drive":
	default:
		return "", fmt.Errorf("invalid backend mode '%v'; expected {proxy, web, caddy, drive}", mode)
	}
	if len(args) != 1 {
		return "", fmt.Errorf("the '%s' backend mode expects a <target>", mode)
	}
	if mode == "proxy" {
		return ParseURL(args[0])
	}
	return args[0], nil
}

// ParseURL accepts a port, host, or HTTP(S) URL using the share command's existing interpretation.
func ParseURL(in string) (string, error) {
	if port, err := strconv.ParseInt(in, 10, 0); err == nil {
		if port <= 0 || port > math.MaxUint16 {
			return "", fmt.Errorf("ports must be between 1 and %d; %d is not", math.MaxUint16, port)
		}
		scheme := "http"
		if port == 443 {
			scheme = "https"
		}
		return fmt.Sprintf("%s://127.0.0.1:%d", scheme, port), nil
	}
	if !strings.HasPrefix(in, "https://") && !strings.HasPrefix(in, "http://") {
		in = "http://" + in
	}
	target, err := url.Parse(in)
	if err != nil {
		return "", err
	}
	return target.String(), nil
}
