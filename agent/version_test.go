package agent

import (
	"context"
	"testing"

	"github.com/openziti/zrok/v2/agent/agentGrpc"
	"github.com/stretchr/testify/require"
)

func TestVersionDeclaresCapabilities(t *testing.T) {
	a := newLifecycleAgent(t)
	version, err := (&agentGrpcImpl{agent: a}).Version(context.Background(), &agentGrpc.VersionRequest{})
	require.NoError(t, err)
	require.True(t, version.HasCapability(agentGrpc.CapabilityBasicAuthFirstColon))
	require.True(t, version.HasCapability(agentGrpc.CapabilityBasicAuthFile))
	require.NotEqual(t, agentGrpc.CapabilityBasicAuthFirstColon, agentGrpc.CapabilityBasicAuthFile)
	require.False(t, (&agentGrpc.VersionResponse{V: "v2.0.9"}).HasCapability(agentGrpc.CapabilityBasicAuthFirstColon))
}
