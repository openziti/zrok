package agent

import (
	"fmt"
	"os"
)

// CommandBuilder provides a fluent interface for building zrok commands
// eliminating repetitive command construction patterns across the agent
type CommandBuilder struct {
	executable string
	args       []string
}

// newCommandBuilder creates a new command builder with the base executable
func newCommandBuilder(subcommands ...string) *CommandBuilder {
	args := []string{os.Args[0]}
	args = append(args, subcommands...)
	return &CommandBuilder{
		executable: os.Args[0],
		args:       args,
	}
}

// NewSharePublicCommand creates a command builder for public share commands
func NewSharePublicCommand() *CommandBuilder {
	return newCommandBuilder("share", "public", "--subordinate")
}

// NewSharePrivateCommand creates a command builder for private share commands
func NewSharePrivateCommand() *CommandBuilder {
	return newCommandBuilder("share", "private", "--subordinate")
}

// NewAccessPrivateCommand creates a command builder for private access commands
func NewAccessPrivateCommand() *CommandBuilder {
	return newCommandBuilder("access", "private", "--subordinate")
}

// BackendMode adds the backend mode option (-b)
func (cb *CommandBuilder) BackendMode(mode string) *CommandBuilder {
	if mode != "" {
		cb.args = append(cb.args, "-b", mode)
	}
	return cb
}

// Target adds the target as the final argument
func (cb *CommandBuilder) Target(target string) *CommandBuilder {
	if target != "" {
		cb.args = append(cb.args, target)
	}
	return cb
}

// ShareToken adds the --share-token option for private shares
func (cb *CommandBuilder) ShareToken(token string) *CommandBuilder {
	if token != "" {
		cb.args = append(cb.args, "--share-token", token)
	}
	return cb
}

// BindAddress adds the bind address as a positional argument for access commands
func (cb *CommandBuilder) BindAddress(address string) *CommandBuilder {
	if address != "" {
		cb.args = append(cb.args, "-b", address)
	}
	return cb
}

// AddFlag adds a simple flag (e.g., --insecure, --open)
func (cb *CommandBuilder) AddFlag(flag string) *CommandBuilder {
	cb.args = append(cb.args, flag)
	return cb
}

// AddOption adds an option with a value (e.g., --option value)
func (cb *CommandBuilder) AddOption(option, value string) *CommandBuilder {
	if value != "" {
		cb.args = append(cb.args, option, value)
	}
	return cb
}

// AddConditionalFlag adds a flag only if the condition is true
func (cb *CommandBuilder) AddConditionalFlag(condition bool, flag string) *CommandBuilder {
	if condition {
		cb.args = append(cb.args, flag)
	}
	return cb
}

// AddMultipleOptions adds multiple instances of the same option
func (cb *CommandBuilder) AddMultipleOptions(option string, values []string) *CommandBuilder {
	for _, value := range values {
		if value != "" {
			cb.args = append(cb.args, option, value)
		}
	}
	return cb
}

// BasicAuth adds multiple --basic-auth options
func (cb *CommandBuilder) BasicAuth(auths []string) *CommandBuilder {
	return cb.AddMultipleOptions("--basic-auth", auths)
}

// NameSelections adds multiple --name-selection options with namespace:name format
func (cb *CommandBuilder) NameSelections(selections []NameSelection) *CommandBuilder {
	for _, nss := range selections {
		nssStr := nss.NamespaceToken
		if nss.Name != "" {
			nssStr += ":" + nss.Name
		}
		if nssStr != "" {
			cb.args = append(cb.args, "--name-selection", nssStr)
		}
	}
	return cb
}

// OauthProvider adds the --oauth-provider option
func (cb *CommandBuilder) OauthProvider(provider string) *CommandBuilder {
	return cb.AddOption("--oauth-provider", provider)
}

// OauthEmailDomains adds multiple --oauth-email-domain options
func (cb *CommandBuilder) OauthEmailDomains(domains []string) *CommandBuilder {
	return cb.AddMultipleOptions("--oauth-email-domain", domains)
}

// OauthRefreshInterval adds the --oauth-refresh-interval option
func (cb *CommandBuilder) OauthRefreshInterval(interval string) *CommandBuilder {
	if interval != "3h0m0s" {
		return cb.AddOption("--oauth-refresh-interval", interval)
	}
	return cb
}

// AccessGrants adds multiple --access-grant options
func (cb *CommandBuilder) AccessGrants(grants []string) *CommandBuilder {
	return cb.AddMultipleOptions("--access-grant", grants)
}

// Insecure adds the --insecure flag if true
func (cb *CommandBuilder) Insecure(insecure bool) *CommandBuilder {
	return cb.AddConditionalFlag(insecure, "--insecure")
}

// Open adds the --open flag if true (note: inverted logic from Closed)
func (cb *CommandBuilder) Open(open bool) *CommandBuilder {
	return cb.AddConditionalFlag(open, "--open")
}

// AutoMode adds auto mode options for access commands
func (cb *CommandBuilder) AutoMode(auto bool, address string, startPort, endPort int) *CommandBuilder {
	if auto {
		cb.args = append(cb.args, "--auto")
		if address != "" {
			cb.args = append(cb.args, "--auto-address", address)
		}
		if startPort > 0 {
			cb.args = append(cb.args, "--auto-start-port", fmt.Sprintf("%d", startPort))
		}
		if endPort > 0 {
			cb.args = append(cb.args, "--auto-end-port", fmt.Sprintf("%d", endPort))
		}
	}
	return cb
}

// Build returns the final command as a string slice
func (cb *CommandBuilder) Build() []string {
	// return a copy to prevent external modification
	result := make([]string, len(cb.args))
	copy(result, cb.args)
	return result
}

// String returns a string representation of the command for debugging
func (cb *CommandBuilder) String() string {
	return fmt.Sprintf("%v", cb.args)
}
