package main

import (
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/openziti/zrok/v2/agent/agentGrpc"
	restEnvironment "github.com/openziti/zrok/v2/rest_client_zrok/environment"
	"github.com/openziti/zrok/v2/rest_client_zrok/metadata"
	"github.com/openziti/zrok/v2/rest_client_zrok/share"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestClassifyFailurePermanentForClientErrors(t *testing.T) {
	for _, err := range []error{
		share.NewShareUnauthorized(),
		share.NewShareNotFound(),
		share.NewUnshareNotFound(),
		share.NewAccessNotFound(),
		metadata.NewGetShareDetailNotFound(),
		restEnvironment.NewDisableUnauthorized(),
		&runtime.APIError{OperationName: "share", Code: http.StatusBadRequest},
	} {
		f := classifyFailure(errors.Wrap(err, "unable to create share"))
		assert.Equal(t, exitPermanent, f.exitCode, "%T", err)
	}
}

func TestClassifyFailureTransientForServerErrors(t *testing.T) {
	for _, err := range []error{
		share.NewShareInternalServerError(),
		share.NewAccessInternalServerError(),
		share.NewShareServiceUnavailable(),
		restEnvironment.NewDisableInternalServerError(),
		&runtime.APIError{OperationName: "share", Code: http.StatusBadGateway},
		&runtime.APIError{OperationName: "share", Code: http.StatusTooManyRequests},
	} {
		f := classifyFailure(errors.Wrap(err, "unable to create share"))
		assert.Equal(t, exitTransient, f.exitCode, "%T", err)
	}
}

func testApiEndpoint() string {
	return "https://api.zrok.example"
}

func TestClassifyFailureTransientForConnectionFailures(t *testing.T) {
	refused := &url.Error{Op: "Post", URL: "https://api.zrok.example/api/v2/share", Err: &net.OpError{Op: "dial", Net: "tcp", Err: syscall.ECONNREFUSED}}
	f := classifyFailureAgainst(errors.Wrap(refused, "unable to create share"), testApiEndpoint)
	assert.Equal(t, exitTransient, f.exitCode)
	assert.Contains(t, f.detail, "unable to reach the zrok controller: ")
}

func TestClassifyFailureTransportFailureElsewhereIsNotUnreachableController(t *testing.T) {
	// the ziti sdk's own requests, made while a backend starts, fail as a *url.Error too.
	zitiRefused := &url.Error{Op: "Post", URL: "https://ziti.zrok.example:1280/edge/client/v1/authenticate", Err: &net.OpError{Op: "dial", Net: "tcp", Err: syscall.ECONNREFUSED}}
	f := classifyFailureAgainst(errors.Wrap(zitiRefused, "error loading ziti context"), testApiEndpoint)
	assert.Equal(t, exitTransient, f.exitCode)
	assert.NotContains(t, f.detail, "unable to reach the zrok controller")
	assert.Equal(t, "error loading ziti context: "+zitiRefused.Error(), f.detail)

	listenErr := &net.OpError{Op: "listen", Net: "ziti", Err: syscall.ECONNRESET}
	f = classifyFailureAgainst(errors.Wrap(listenErr, "error listening"), testApiEndpoint)
	assert.Equal(t, exitTransient, f.exitCode)
	assert.Equal(t, "error listening: "+listenErr.Error(), f.detail)
}

func TestClassifyFailurePlainErrorIsTransient(t *testing.T) {
	f := classifyFailureAgainst(errors.New("invalid backend mode"), testApiEndpoint)
	assert.Equal(t, exitTransient, f.exitCode)
	assert.Equal(t, "invalid backend mode", f.detail)
	assert.Equal(t, "unable to create share (invalid backend mode)", f.message("unable to create share"))
}

func TestClassifyFailureBackendFileErrorPrintsAsItself(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "Caddyfile")
	_, err := os.ReadFile(missing)
	require.Error(t, err)

	f := classifyFailureAgainst(err, testApiEndpoint)
	assert.Equal(t, exitTransient, f.exitCode)
	assert.Equal(t, "open "+missing+": no such file or directory", f.detail)
	assert.Equal(t, "unable to create caddy backend (open "+missing+": no such file or directory)", f.message("unable to create caddy backend"))
}

func TestClassifyFailurePermanentErrorIsPermanent(t *testing.T) {
	f := classifyFailure(newPermanentError("name '%v' is not held by a live share", "x"))
	assert.Equal(t, exitPermanent, f.exitCode)
	assert.Equal(t, "name 'x' is not held by a live share", f.detail)
}

func TestFailureMessageUsesConflictPayloadAndAdvice(t *testing.T) {
	conflict := &share.ShareConflict{Payload: rest_model_zrok.ErrorMessage("name 'x' in namespace 'public' is held by share 'abc123'")}
	f := classifyFailure(errors.Wrap(conflict, "unable to create share"))
	assert.Equal(t, exitPermanent, f.exitCode)
	assert.Equal(t, "unable to create share (name 'x' in namespace 'public' is held by share 'abc123'); "+
		"retrying will not change the answer; 'zrok2 agent' retries transient failures with backoff", f.message("unable to create share"))
	assert.Equal(t, "unable to create share (name 'x' in namespace 'public' is held by share 'abc123')", f.describe("unable to create share"))
}

func TestFailureMessageWithoutPayloadNamesStatus(t *testing.T) {
	f := classifyFailure(share.NewShareUnauthorized())
	assert.Equal(t, "the zrok controller answered '401 unauthorized'; check the account token", f.detail)
}

func TestFailureMessageWithoutPayloadHints(t *testing.T) {
	assert.Equal(t, "the zrok controller answered '404 not found'; check the token or resource", classifyFailure(share.NewShareNotFound()).detail)
	assert.Equal(t, "the zrok controller answered '409 conflict'", classifyFailure(&runtime.APIError{OperationName: "share", Code: http.StatusConflict}).detail)
}

func TestFailureMessageTransientHasNoAdvice(t *testing.T) {
	f := classifyFailure(share.NewShareInternalServerError())
	assert.Equal(t, "unable to create share (the zrok controller answered '500 internal server error')", f.message("unable to create share"))
}

func TestServiceUnavailableMessageUsesRetryAfter(t *testing.T) {
	for _, err := range []error{
		&share.ShareServiceUnavailable{RetryAfter: 5},
		&share.UnshareServiceUnavailable{RetryAfter: 5},
		&share.AccessServiceUnavailable{RetryAfter: 5},
		&share.UnaccessServiceUnavailable{RetryAfter: 5},
		&restEnvironment.EnableServiceUnavailable{RetryAfter: 5},
		&restEnvironment.DisableServiceUnavailable{RetryAfter: 5},
	} {
		f := classifyFailure(errors.Wrap(err, "unable to create share"))
		assert.Equal(t, exitTransient, f.exitCode, "%T", err)
		assert.Equal(t, "the zrok service is busy, retry in 5 seconds", f.detail, "%T", err)
	}
}

func TestBusyMessage(t *testing.T) {
	assert.Equal(t, "the zrok service is busy, retry in 30 seconds", busyMessage(30))
	assert.Equal(t, "the zrok service is busy, retry in 1 second", busyMessage(1))
	assert.Equal(t, "the zrok service is busy, retry shortly", busyMessage(0))
}

func TestClassifyAgentControllerResponse(t *testing.T) {
	for _, test := range []struct {
		response agentGrpc.ControllerResponse
		code     int
		detail   string
	}{
		{agentGrpc.ControllerResponse{Status: 400, Payload: "client upgrade required"}, exitPermanent, "client upgrade required"},
		{agentGrpc.ControllerResponse{Status: 409, Payload: "name 'app' is held by share 'existing'"}, exitPermanent, "name 'app' is held by share 'existing'"},
		{agentGrpc.ControllerResponse{Status: 401}, exitPermanent, "the zrok controller answered '401 unauthorized'; check the account token"},
		{agentGrpc.ControllerResponse{Status: 429}, exitTransient, "the zrok service is busy, retry shortly"},
		{agentGrpc.ControllerResponse{Status: 500, Payload: "controller failure"}, exitTransient, "controller failure"},
		{agentGrpc.ControllerResponse{Status: 503, RetryAfter: 5}, exitTransient, "the zrok service is busy, retry in 5 seconds"},
	} {
		rpcError := test.response.AsError(errors.New("controller request failed"))
		// classification receives a reconstructed wire status, without the server's Go error types.
		wireStatus := status.Convert(rpcError).Proto()
		f := classifyFailure(status.FromProto(wireStatus).Err())
		require.Equal(t, test.code, f.exitCode)
		require.Equal(t, test.detail, f.detail)
	}
	for _, err := range []error{
		status.Error(codes.Unknown, "controller answered 409"),
		status.Error(codes.Unavailable, "agent unavailable"),
	} {
		require.Equal(t, exitTransient, classifyFailure(err).exitCode, "unstructured RPC errors must stay transient")
	}
}

func TestVersionCheckRejectionPreservesCLIMessage(t *testing.T) {
	err := errors.Wrap(&metadata.ClientVersionCheckBadRequest{Payload: "client upgrade required"}, "client initialization failed")
	f := classifyFailure(err)
	require.Equal(t, exitPermanent, f.exitCode)
	require.Equal(t, "client upgrade required", f.detail)
}
