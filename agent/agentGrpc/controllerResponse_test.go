package agentGrpc

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestControllerResponseMetadataWireCompatibility(t *testing.T) {
	response := ControllerResponse{Status: 503, Payload: "controller busy", RetryAfter: 5}
	err := response.AsError(errors.New("create failed"))
	rpc := status.Convert(err)
	require.Equal(t, "create failed", rpc.Message())
	require.Len(t, rpc.Details(), 1)
	detail, ok := rpc.Details()[0].(*errdetails.ErrorInfo)
	require.True(t, ok)
	require.Equal(t, "zrok.agent", detail.Domain)
	require.Equal(t, "CONTROLLER_RESPONSE", detail.Reason)
	require.Equal(t, map[string]string{"http_status": "503", "payload": "controller busy", "retry_after": "5"}, detail.Metadata)

	// decode a fixed wire fixture independently of the encoder's field binding.
	legacy, err := status.New(codes.Unknown, "refused").WithDetails(&errdetails.ErrorInfo{
		Domain: "zrok.agent", Reason: "CONTROLLER_RESPONSE",
		Metadata: map[string]string{"http_status": "409", "payload": "name 'app' is taken", "retry_after": "0"},
	})
	require.NoError(t, err)
	decoded, ok := ControllerResponseFromError(legacy.Err())
	require.True(t, ok)
	require.Equal(t, ControllerResponse{Status: 409, Payload: "name 'app' is taken"}, decoded)
}

func TestControllerResponseMetadataRequiresControllerStatus(t *testing.T) {
	for _, metadata := range []map[string]string{
		nil, {"payload": "refused"}, {"http_status": "invalid"}, {"http_status": "200"}, {"http_status": "600"},
	} {
		rpc, err := status.New(codes.Unknown, "unknown").WithDetails(&errdetails.ErrorInfo{
			Domain: "zrok.agent", Reason: "CONTROLLER_RESPONSE", Metadata: metadata,
		})
		require.NoError(t, err)
		_, ok := ControllerResponseFromError(rpc.Err())
		require.False(t, ok, "%v", metadata)
	}
}
