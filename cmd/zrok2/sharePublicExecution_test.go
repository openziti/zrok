package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/openziti/zrok/v2/agent/agentGrpc"
	"github.com/openziti/zrok/v2/environment"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestPublicOptionExecutionHelper(t *testing.T) {
	if os.Getenv("ZROK_PUBLIC_OPTIONS_HELPER") != "1" {
		return
	}
	environment.SetRootDirName(os.Getenv("ZROK_PUBLIC_OPTIONS_ROOT"))
	command := newSharePublicCommand()
	for index, arg := range os.Args {
		if arg == "--" {
			require.NoError(t, command.cmd.ParseFlags(os.Args[index+1:]))
			command.run(command.cmd, []string{"8080"})
			t.Fatal("invalid worker invocation returned successfully")
		}
	}
	t.Fatal("missing helper arguments")
}

func TestPublicCLIRejectsInvalidWorkerInvocationBeforeRouting(t *testing.T) {
	root := t.TempDir()
	var controllerCalls, agentCalls atomic.Int32
	controller := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		controllerCalls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer controller.Close()
	require.NoError(t, os.WriteFile(filepath.Join(root, "metadata.json"), []byte(`{"v":"v0.4"}`), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "environment.json"), []byte(fmt.Sprintf(`{"zrok_token":"account","ziti_identity":"identity","api_endpoint":%q}`, controller.URL)), 0600))
	listener, err := net.Listen("unix", filepath.Join(root, "agent.socket"))
	require.NoError(t, err)
	server := grpc.NewServer(grpc.UnaryInterceptor(func(ctx context.Context, request any, _ *grpc.UnaryServerInfo, next grpc.UnaryHandler) (any, error) {
		agentCalls.Add(1)
		return next(ctx, request)
	}))
	agentGrpc.RegisterAgentServer(server, &agentGrpc.UnimplementedAgentServer{})
	go func() { _ = server.Serve(listener) }()
	defer server.Stop()
	defer listener.Close()
	for _, flags := range [][]string{
		{"--share-token", "owned"}, {"--share-token="}, {"--share-token", " "},
		{"--share-token", "owned", "--force-agent"}, {"--share-token", "owned", "--force-local"},
		{"--subordinate"}, {"--subordinate", "--share-token="}, {"--subordinate", "--share-token", "  "},
	} {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		args := append([]string{"-test.run=^TestPublicOptionExecutionHelper$", "--"}, flags...)
		child := exec.CommandContext(ctx, os.Args[0], args...)
		child.Env = append(os.Environ(), "ZROK_PUBLIC_OPTIONS_HELPER=1", "ZROK_PUBLIC_OPTIONS_ROOT="+root)
		output, err := child.CombinedOutput()
		cancel()
		var exited *exec.ExitError
		require.ErrorAs(t, err, &exited, "%s", output)
		require.Equal(t, exitPermanent, exited.ExitCode(), "%s", output)
		require.Contains(t, string(output), "invalid public worker options")
		require.Zero(t, controllerCalls.Load(), "invalid invocation reached the controller")
		require.Zero(t, agentCalls.Load(), "invalid invocation attempted agent detection or routing")
	}
	// even an unusable environment must not supersede option validation.
	require.NoError(t, os.WriteFile(filepath.Join(root, "metadata.json"), []byte("{"), 0600))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	child := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestPublicOptionExecutionHelper$", "--", "--subordinate")
	child.Env = append(os.Environ(), "ZROK_PUBLIC_OPTIONS_HELPER=1", "ZROK_PUBLIC_OPTIONS_ROOT="+root)
	output, err := child.CombinedOutput()
	require.Error(t, err)
	require.Contains(t, string(output), "invalid public worker options")
	require.NotContains(t, string(output), "unable to load environment")
}

// legacyAgent answers as an agent from an earlier release of the series: it declares no capabilities, and it
// rejects a request by quoting the credential it received, as v2.0.9 does.
type legacyAgent struct {
	agentGrpc.UnimplementedAgentServer
	shares atomic.Int32
}

func (a *legacyAgent) Version(context.Context, *agentGrpc.VersionRequest) (*agentGrpc.VersionResponse, error) {
	return &agentGrpc.VersionResponse{V: "v2.0.9"}, nil
}

func (a *legacyAgent) SharePublic(_ context.Context, req *agentGrpc.SharePublicRequest) (*agentGrpc.SharePublicResponse, error) {
	a.shares.Add(1)
	return nil, fmt.Errorf("invalid username:password '%v'", strings.Join(req.BasicAuth, ","))
}

// publicExecutionFixture is an enabled environment whose controller and agent count every call they receive.
type publicExecutionFixture struct {
	root            string
	controllerCalls atomic.Int32
	agent           *legacyAgent
}

func newPublicExecutionFixture(t *testing.T) *publicExecutionFixture {
	f := &publicExecutionFixture{root: t.TempDir(), agent: &legacyAgent{}}
	controller := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		f.controllerCalls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(controller.Close)
	require.NoError(t, os.WriteFile(filepath.Join(f.root, "metadata.json"), []byte(`{"v":"v0.4"}`), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(f.root, "environment.json"), []byte(fmt.Sprintf(`{"zrok_token":"account","ziti_identity":"identity","api_endpoint":%q}`, controller.URL)), 0600))
	listener, err := net.Listen("unix", filepath.Join(f.root, "agent.socket"))
	require.NoError(t, err)
	server := grpc.NewServer()
	agentGrpc.RegisterAgentServer(server, f.agent)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() {
		server.Stop()
		_ = listener.Close()
	})
	return f
}

// run executes 'share public 8080' with flags in a child process and returns its exit status and output.
func (f *publicExecutionFixture) run(t *testing.T, stdin io.Reader, flags ...string) (int, string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	child := exec.CommandContext(ctx, os.Args[0], append([]string{"-test.run=^TestPublicOptionExecutionHelper$", "--"}, flags...)...)
	child.Env = append(os.Environ(), "ZROK_PUBLIC_OPTIONS_HELPER=1", "ZROK_PUBLIC_OPTIONS_ROOT="+f.root)
	child.Stdin = stdin
	output, err := child.CombinedOutput()
	var exited *exec.ExitError
	require.ErrorAs(t, err, &exited, "%s", output)
	return exited.ExitCode(), string(output)
}

func TestPublicCLIWorkerRefusesStdinBeforeReadingIt(t *testing.T) {
	f := newPublicExecutionFixture(t)
	// stdin stays open and empty; a read before validation would block until the child is killed.
	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	defer reader.Close()
	defer writer.Close()
	code, output := f.run(t, reader, "--subordinate", "--share-token", "owned", "--basic-auth-from-stdin")
	require.Equal(t, exitPermanent, code, output)
	require.Contains(t, output, "invalid public worker options")
	require.Contains(t, output, "--basic-auth-from-stdin")
	require.Zero(t, f.controllerCalls.Load())
	require.Zero(t, f.agent.shares.Load())
}

func TestPublicCLIMalformedBasicAuthIsNotQuoted(t *testing.T) {
	f := newPublicExecutionFixture(t)
	for _, tc := range []struct {
		stdin string
		flags []string
	}{
		{flags: []string{"--force-local", "--basic-auth", "admin:secret", "--basic-auth", basicAuthSentinel}},
		{flags: []string{"--force-agent", "--basic-auth", basicAuthSentinel}},
		{stdin: basicAuthSentinel + "\n", flags: []string{"--force-local", "--basic-auth-from-stdin"}},
		{stdin: basicAuthSentinel + "\n", flags: []string{"--basic-auth-from-stdin"}},
	} {
		code, output := f.run(t, strings.NewReader(tc.stdin), tc.flags...)
		require.Equal(t, exitPermanent, code, "%v: %s", tc.flags, output)
		require.Contains(t, output, "invalid basic authentication", "%v", tc.flags)
		require.NotContains(t, output, basicAuthSentinel, "%v", tc.flags)
	}
	require.Zero(t, f.controllerCalls.Load())
	require.Zero(t, f.agent.shares.Load())
}

func TestPublicCLIRefusesColonPasswordForLegacyAgent(t *testing.T) {
	f := newPublicExecutionFixture(t)
	credential := "admin:" + basicAuthSentinel + ":x"
	for _, tc := range []struct {
		stdin string
		flags []string
	}{
		{flags: []string{"--basic-auth", credential}},
		{flags: []string{"--force-agent", "--basic-auth", credential}},
		{stdin: credential + "\n", flags: []string{"--basic-auth-from-stdin"}},
		{stdin: credential + "\n", flags: []string{"--force-agent", "--basic-auth-from-stdin"}},
	} {
		code, output := f.run(t, strings.NewReader(tc.stdin), tc.flags...)
		require.Equal(t, exitPermanent, code, "%v: %s", tc.flags, output)
		require.Contains(t, output, "restart 'zrok2 agent'", "%v", tc.flags)
		require.Contains(t, output, "v2.0.9", "%v", tc.flags)
		require.NotContains(t, output, basicAuthSentinel, "%v", tc.flags)
		require.Zero(t, f.agent.shares.Load(), "%v: the credential was sent to an agent that cannot take it", tc.flags)
	}
	require.Zero(t, f.controllerCalls.Load())

	// an entry the earlier grammar accepts goes to any agent, as it always has.
	_, output := f.run(t, nil, "--basic-auth", "admin:secret")
	require.EqualValues(t, 1, f.agent.shares.Load(), output)
}
