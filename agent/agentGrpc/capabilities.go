package agentGrpc

import "slices"

// an agent declares in its VersionResponse what it understands beyond its release series. the command checks
// a capability before sending a request that an agent from an earlier release of the same series would mishandle.
const (
	// CapabilityBasicAuthFirstColon: basic authentication entries are split at the first colon, so a password
	// may contain one, and a rejected entry is never quoted in an error.
	CapabilityBasicAuthFirstColon = "basic-auth-first-colon"
)

// HasCapability reports whether the agent's VersionResponse declares capability.
func (x *VersionResponse) HasCapability(capability string) bool {
	return slices.Contains(x.GetCapabilities(), capability)
}
