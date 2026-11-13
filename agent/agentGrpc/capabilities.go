package agentGrpc

import "slices"

// an agent declares in its VersionResponse what it understands beyond its release series. the command checks
// a capability before sending a request that an agent from an earlier release of the same series would mishandle.
const (
	// CapabilityBasicAuthFirstColon: basic authentication entries are split at the first colon, so a password
	// may contain one, and a rejected entry is never quoted in an error.
	CapabilityBasicAuthFirstColon = "basic-auth-first-colon"

	// CapabilityBasicAuthFile: a SharePublicRequest's basicAuthFile is a reference the agent reads at each
	// create. an agent without it ignores the field and would create the share with no authentication.
	CapabilityBasicAuthFile = "basic-auth-file"
)

// HasCapability reports whether the agent's VersionResponse declares capability.
func (x *VersionResponse) HasCapability(capability string) bool {
	return slices.Contains(x.GetCapabilities(), capability)
}
