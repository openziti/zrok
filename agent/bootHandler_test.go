package agent

import (
	"testing"

	"github.com/openziti/zrok/v2/cmd/zrok2/subordinate"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
	"github.com/stretchr/testify/require"
)

func TestPublicWorkerBootValidatesWithoutReplacingParentFields(t *testing.T) {
	for _, token := range []interface{}{nil, "", " ", "other", 7, "owned"} {
		shr := &share{token: "owned", frontendEndpoints: []string{"parent-endpoint"}, target: "parent-target", backendMode: sdk.ProxyBackendMode, shareMode: sdk.PublicShareMode}
		var bootErr error
		handler := NewPublicWorkerBootHandler(shr, &bootErr)
		msg := subordinate.Message{"token": token, "frontend_endpoints": []interface{}{"child-endpoint"}, "target": "child-target", "backend_mode": "drive", "share_mode": "private"}
		handler.HandleBoot(subordinate.BootMessage, msg)
		if token == "owned" {
			require.NoError(t, bootErr)
		} else {
			require.Error(t, bootErr)
		}
		require.Equal(t, "owned", shr.token)
		require.Equal(t, []string{"parent-endpoint"}, shr.frontendEndpoints)
		require.Equal(t, "parent-target", shr.target)
		require.Equal(t, sdk.ProxyBackendMode, shr.backendMode)
		require.Equal(t, sdk.PublicShareMode, shr.shareMode)
	}
	var bootErr error
	NewPublicWorkerBootHandler(&share{}, &bootErr).HandleBoot(subordinate.BootMessage, subordinate.Message{"token": "owned"})
	require.Error(t, bootErr)
}

func TestLegacyBootFieldsRemainDistinct(t *testing.T) {
	var bootErr error
	shr := &share{}
	NewShareBootHandler(shr, &bootErr).HandleBoot(subordinate.BootMessage, subordinate.Message{"token": "private-token", "backend_mode": "tcpTunnel", "share_mode": "private", "target": "host:80", "frontend_endpoints": []interface{}{"endpoint"}})
	require.NoError(t, bootErr)
	require.Equal(t, "private-token", shr.token)
	require.Equal(t, sdk.TcpTunnelBackendMode, shr.backendMode)
	require.Equal(t, sdk.PrivateShareMode, shr.shareMode)
	require.Equal(t, "host:80", shr.target)
	require.Equal(t, []string{"endpoint"}, shr.frontendEndpoints)
	acc := &access{}
	NewAccessBootHandler(acc, &bootErr).HandleBoot(subordinate.BootMessage, subordinate.Message{"frontend_token": "frontend", "bind_address": "127.0.0.1:1234"})
	require.NoError(t, bootErr)
	require.Equal(t, "frontend", acc.frontendToken)
	require.Equal(t, "127.0.0.1:1234", acc.bindAddress)
}
