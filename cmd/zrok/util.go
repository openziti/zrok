package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti/zrok/agent/agentClient"
	"github.com/openziti/zrok/cmd/zrok/subordinate"
	"github.com/openziti/zrok/endpoints/vpn"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/environment/env_core"
	"github.com/openziti/zrok/tui"
	"github.com/pkg/errors"
)

func mustGetAdminAuth() runtime.ClientAuthInfoWriter {
	adminToken := os.Getenv("ZROK_ADMIN_TOKEN")
	if adminToken == "" {
		panic("please set ZROK_ADMIN_TOKEN to a valid admin token for your zrok instance")
	}
	return httptransport.APIKeyAuth("X-TOKEN", "header", adminToken)
}

func mustGetEnvironmentAuth() (env_core.Root, runtime.ClientAuthInfoWriter) {
	env, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}
	if !env.IsEnabled() {
		panic("environment is not enabled; run 'zrok enable' first")
	}
	auth := httptransport.APIKeyAuth("X-TOKEN", "header", env.Environment().AccountToken)
	return env, auth
}

func parseUrl(in string) (string, error) {
	// parse port-only urls
	if iv, err := strconv.ParseInt(in, 10, 0); err == nil {
		if iv > 0 && iv <= math.MaxUint16 {
			if iv == 443 {
				return fmt.Sprintf("https://127.0.0.1:%d", iv), nil
			}
			return fmt.Sprintf("http://127.0.0.1:%d", iv), nil
		}
		return "", errors.Errorf("ports must be between 1 and %d; %d is not", math.MaxUint16, iv)
	}

	// make sure either https:// or http:// was specified
	if !strings.HasPrefix(in, "https://") && !strings.HasPrefix(in, "http://") {
		in = "http://" + in
	}

	// parse the url
	targetEndpoint, err := url.Parse(in)
	if err != nil {
		return "", err
	}

	return targetEndpoint.String(), nil
}

func subordinateError(err error) {
	msg := make(map[string]interface{})
	msg[subordinate.MessageKey] = subordinate.ErrorMessage
	msg[subordinate.ErrorMessage] = err.Error()
	if data, err := json.Marshal(msg); err == nil {
		fmt.Println(string(data))
	} else {
		fmt.Println("{\"" + subordinate.MessageKey + "\":\"" + subordinate.ErrorMessage + "\",\"" + subordinate.ErrorMessage + "\":\"internal error\"}")
	}
	os.Exit(1)
}

// detectAndRouteToAgent handles the common pattern of checking if the agent is running
// and routing to either agent or local execution paths. This eliminates duplicate code
// found in sharePrivate, sharePublic, and accessPrivate commands.
func detectAndRouteToAgent(
	subordinate, forceLocal, forceAgent bool,
	root env_core.Root,
	localFn func(),
	agentFn func(),
) {
	// if running in subordinate mode or forced local, always use local
	if subordinate || forceLocal {
		localFn()
		return
	}

	// determine if agent is running
	agent := forceAgent
	if !forceAgent {
		var err error
		agent, err = agentClient.IsAgentRunning(root)
		if err != nil {
			tui.Error("error checking if agent is running", err)
		}
	}

	// route to appropriate handler
	if agent {
		agentFn()
	} else {
		localFn()
	}
}

// backendModeConfig holds the configuration for validating and processing backend modes
type backendModeConfig struct {
	expectsTarget bool
	parseTarget   func(string) (string, error)
	forceHeadless bool
}

// validateBackendMode validates the backend mode and processes the target argument.
// This eliminates the duplicate switch statements found across share commands.
// Returns the processed target string and whether headless mode should be forced.
// Set allowedModes to nil to allow all backend modes, or provide a list to restrict.
func validateBackendMode(mode string, args []string, allowedModes []string) (target string, forceHeadless bool, err error) {
	configs := map[string]backendModeConfig{
		"proxy": {
			expectsTarget: true,
			parseTarget:   parseUrl,
			forceHeadless: false,
		},
		"web": {
			expectsTarget: true,
			parseTarget:   func(s string) (string, error) { return s, nil },
			forceHeadless: false,
		},
		"tcpTunnel": {
			expectsTarget: true,
			parseTarget:   func(s string) (string, error) { return s, nil },
			forceHeadless: false,
		},
		"udpTunnel": {
			expectsTarget: true,
			parseTarget:   func(s string) (string, error) { return s, nil },
			forceHeadless: false,
		},
		"caddy": {
			expectsTarget: true,
			parseTarget:   func(s string) (string, error) { return s, nil },
			forceHeadless: true,
		},
		"drive": {
			expectsTarget: true,
			parseTarget:   func(s string) (string, error) { return s, nil },
			forceHeadless: false,
		},
		"socks": {
			expectsTarget: false,
			parseTarget:   nil,
			forceHeadless: false,
		},
		"vpn": {
			expectsTarget: false, // vpn is optional - can use default
			parseTarget: func(s string) (string, error) {
				if s == "" {
					return vpn.DefaultTarget(), nil
				}
				_, _, err := net.ParseCIDR(s)
				if err != nil {
					return "", errors.New("the 'vpn' backend mode expects a valid CIDR <target>")
				}
				return s, nil
			},
			forceHeadless: false,
		},
	}

	// check if mode is allowed
	if allowedModes != nil {
		allowed := false
		for _, m := range allowedModes {
			if m == mode {
				allowed = true
				break
			}
		}
		if !allowed {
			return "", false, fmt.Errorf("invalid backend mode '%v'; expected {%s}", mode, strings.Join(allowedModes, ", "))
		}
	}

	config, ok := configs[mode]
	if !ok {
		// build list of valid modes - either from allowedModes or all available
		validModes := allowedModes
		if validModes == nil {
			validModes = make([]string, 0, len(configs))
			for k := range configs {
				validModes = append(validModes, k)
			}
		}
		return "", false, fmt.Errorf("invalid backend mode '%v'; expected {%s}", mode, strings.Join(validModes, ", "))
	}

	// handle special cases
	switch mode {
	case "socks":
		// socks doesn't expect arguments
		if len(args) != 0 {
			return "", false, errors.New("the 'socks' backend mode does not expect a <target>")
		}
		return "socks", config.forceHeadless, nil

	case "vpn":
		// vpn has optional target
		if len(args) == 0 {
			return vpn.DefaultTarget(), config.forceHeadless, nil
		} else if len(args) == 1 {
			target, err = config.parseTarget(args[0])
			if err != nil {
				return "", false, errors.Wrap(err, "unable to create share")
			}
			return target, config.forceHeadless, nil
		} else {
			return "", false, errors.New("the 'vpn' backend mode expects at most one <target>")
		}

	default:
		// standard modes that expect exactly one target
		if config.expectsTarget {
			if len(args) != 1 {
				return "", false, fmt.Errorf("the '%s' backend mode expects a <target>", mode)
			}

			target, err = config.parseTarget(args[0])
			if err != nil {
				if mode == "proxy" {
					return "", false, errors.Wrap(err, "invalid target endpoint URL")
				}
				return "", false, errors.Wrapf(err, "invalid target for backend mode '%s'", mode)
			}
			return target, config.forceHeadless, nil
		}
	}

	return "", false, fmt.Errorf("unexpected backend mode configuration for '%s'", mode)
}

func (cmd *agentStatusCommand) wrapString(s string, maxWidth int) string {
	if len(s) <= maxWidth {
		return s
	}

	var result []rune
	line := []rune{}
	words := [][]rune{}
	currentWord := []rune{}

	// split input into words
	for _, r := range s {
		if r == ' ' || r == '\t' || r == '\n' {
			if len(currentWord) > 0 {
				words = append(words, currentWord)
				currentWord = []rune{}
			}
			if r == '\n' {
				// preserve existing newlines
				words = append(words, []rune{r})
			}
		} else {
			currentWord = append(currentWord, r)
		}
	}
	if len(currentWord) > 0 {
		words = append(words, currentWord)
	}

	// wrap words into lines
	for _, word := range words {
		if len(word) == 1 && word[0] == '\n' {
			// handle preserved newlines
			result = append(result, line...)
			result = append(result, '\n')
			line = []rune{}
			continue
		}

		// check if adding this word would exceed the width
		spaceNeeded := 0
		if len(line) > 0 {
			spaceNeeded = 1 // for the space between words
		}

		if len(line)+spaceNeeded+len(word) > maxWidth {
			// word doesn't fit on current line
			if len(line) > 0 {
				// flush current line
				result = append(result, line...)
				result = append(result, '\n')
				line = []rune{}
			}

			// if word itself is longer than maxWidth, break it
			if len(word) > maxWidth {
				for i := 0; i < len(word); {
					end := i + maxWidth
					if end > len(word) {
						end = len(word)
					}
					if i > 0 {
						result = append(result, '\n')
					}
					result = append(result, word[i:end]...)
					i = end
				}
				if len(word) > 0 && len(word)%maxWidth != 0 {
					result = append(result, '\n')
				}
			} else {
				// word fits on new line
				line = append(line, word...)
			}
		} else {
			// word fits on current line
			if len(line) > 0 {
				line = append(line, ' ')
			}
			line = append(line, word...)
		}
	}

	// append any remaining line content
	if len(line) > 0 {
		result = append(result, line...)
	}

	return string(result)
}
