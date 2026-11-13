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
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/openziti/zrok/v2/agent/agentGrpc"
	"github.com/openziti/zrok/v2/environment"
	"github.com/openziti/zrok/v2/tui"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func TestPublicOptionExecutionHelper(t *testing.T) {
	if os.Getenv("ZROK_PUBLIC_OPTIONS_HELPER") != "1" {
		return
	}
	environment.SetRootDirName(os.Getenv("ZROK_PUBLIC_OPTIONS_ROOT"))
	command := newSharePublicCommand()
	for index, arg := range os.Args {
		if arg == "--" {
			// execute through cobra, as the binary does, so flag-group validation runs; report its error as main does.
			command.cmd.SetArgs(append(append([]string{}, os.Args[index+1:]...), "8080"))
			if err := command.cmd.Execute(); err != nil {
				tui.Error("an error occurred", err)
			}
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

// legacyAgent answers as an agent from an earlier release of the series: by default it declares no capabilities,
// and it rejects a request by quoting the credential it received, as v2.0.9 does. it records every request.
type legacyAgent struct {
	agentGrpc.UnimplementedAgentServer
	capabilities []string
	refusal      error // when set, SharePublic answers with it instead of quoting
	versions     atomic.Int32
	shares       atomic.Int32
	mu           sync.Mutex
	requests     []*agentGrpc.SharePublicRequest
}

func (a *legacyAgent) Version(context.Context, *agentGrpc.VersionRequest) (*agentGrpc.VersionResponse, error) {
	a.versions.Add(1)
	a.mu.Lock()
	defer a.mu.Unlock()
	return &agentGrpc.VersionResponse{V: "v2.0.9", Capabilities: slices.Clone(a.capabilities)}, nil
}

func (a *legacyAgent) declare(capabilities ...string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.capabilities = capabilities
}

func (a *legacyAgent) SharePublic(_ context.Context, req *agentGrpc.SharePublicRequest) (*agentGrpc.SharePublicResponse, error) {
	a.shares.Add(1)
	a.mu.Lock()
	a.requests = append(a.requests, proto.Clone(req).(*agentGrpc.SharePublicRequest))
	refusal := a.refusal
	a.mu.Unlock()
	if refusal != nil {
		return nil, refusal
	}
	return nil, fmt.Errorf("invalid username:password '%v'", strings.Join(req.BasicAuth, ","))
}

func (a *legacyAgent) received() []*agentGrpc.SharePublicRequest {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]*agentGrpc.SharePublicRequest(nil), a.requests...)
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

func TestPublicCLIBasicAuthFileRefusals(t *testing.T) {
	f := newPublicExecutionFixture(t)
	dir := t.TempDir()
	write := func(name, content string) string {
		path := filepath.Join(dir, name)
		require.NoError(t, os.WriteFile(path, []byte(content), 0600))
		return path
	}
	good := write("good", "admin:s3:cret\n")
	malformed := write("malformed", "admin:one\n"+basicAuthSentinel+"\n")
	empty := write("empty", "\n \n")

	t.Run("an explicitly empty path exits 2 before routing", func(t *testing.T) {
		for _, flags := range [][]string{
			{"--force-local", "--basic-auth-file="}, {"--basic-auth-file", ""}, {"--force-agent", "--basic-auth-file", "  "},
		} {
			code, output := f.run(t, nil, flags...)
			require.Equal(t, exitPermanent, code, "%v: %s", flags, output)
			require.Contains(t, output, "--basic-auth-file requires a nonempty path", "%v", flags)
		}
		require.Zero(t, f.controllerCalls.Load())
		require.Zero(t, f.agent.versions.Load(), "an empty path reached agent detection")
		require.Zero(t, f.agent.shares.Load())
	})

	t.Run("a bad file exits 2 with no create", func(t *testing.T) {
		for _, path := range []string{malformed, empty, filepath.Join(dir, "missing"), dir} {
			code, output := f.run(t, nil, "--force-local", "--basic-auth-file", path)
			require.Equal(t, exitPermanent, code, "%v: %s", path, output)
			require.Contains(t, output, path)
			require.NotContains(t, output, basicAuthSentinel)
		}
		require.Zero(t, f.controllerCalls.Load())
		require.Zero(t, f.agent.shares.Load())
	})

	t.Run("an agent without the file capability is sent nothing", func(t *testing.T) {
		// with no capabilities (v2.0.9), and with only the entry grammar (a stage 1 build).
		for _, capabilities := range [][]string{nil, {agentGrpc.CapabilityBasicAuthFirstColon}} {
			f.agent.declare(capabilities...)
			for _, flags := range [][]string{
				{"--basic-auth-file", good}, {"--force-agent", "--basic-auth-file", good}, {"--force-agent", "--basic-auth-file", malformed},
			} {
				code, output := f.run(t, nil, flags...)
				require.Equal(t, exitPermanent, code, "%v: %s", flags, output)
				require.Contains(t, output, "restart 'zrok2 agent'", "%v", flags)
				require.Contains(t, output, "v2.0.9", "%v", flags)
				require.NotContains(t, output, basicAuthSentinel)
			}
		}
		f.agent.declare()
		require.Zero(t, f.controllerCalls.Load())
		require.Zero(t, f.agent.shares.Load(), "a file reference was sent to an agent that would ignore it")
	})

	t.Run("exclusive sources are rejected", func(t *testing.T) {
		for _, flags := range [][]string{
			{"--basic-auth-file", good, "--basic-auth", "admin:secret"},
			{"--basic-auth-file", good, "--basic-auth-from-stdin"},
			{"--basic-auth-file", good, "--oauth-provider", "github"},
			{"--force-local", "--basic-auth-file=", "--basic-auth", "admin:secret"},
		} {
			code, output := f.run(t, strings.NewReader("admin:secret\n"), flags...)
			require.NotZero(t, code, "%v: %s", flags, output)
			require.Contains(t, output, "if any flags in the group", "%v", flags)
		}
		require.Zero(t, f.controllerCalls.Load())
		require.Zero(t, f.agent.shares.Load())
	})
}

func TestPublicCLISendsBasicAuthFileByReference(t *testing.T) {
	f := newPublicExecutionFixture(t)
	f.agent.declare(agentGrpc.CapabilityBasicAuthFirstColon, agentGrpc.CapabilityBasicAuthFile)
	cwd, err := os.Getwd()
	require.NoError(t, err)
	// the paths do not exist: the agent is the only reader, so the command must not try.
	absent := filepath.Join(t.TempDir(), "absent-users")
	for _, tc := range []struct {
		path string
		want string
	}{
		{absent, absent},
		{"relative-absent-users", filepath.Join(cwd, "relative-absent-users")},
	} {
		for _, force := range []bool{false, true} {
			flags := []string{"--basic-auth-file", tc.path}
			if force {
				flags = append([]string{"--force-agent"}, flags...)
			}
			before := len(f.agent.received())
			_, output := f.run(t, nil, flags...)
			received := f.agent.received()
			require.Len(t, received, before+1, "%v: %s", flags, output)
			sent := received[len(received)-1]
			require.Equal(t, tc.want, sent.BasicAuthFile, "%v", flags)
			require.Empty(t, sent.BasicAuth, "%v", flags)
		}
	}
	require.Zero(t, f.controllerCalls.Load())
}

func TestPublicCLIAgentFileRefusalExitsPermanent(t *testing.T) {
	f := newPublicExecutionFixture(t)
	f.agent.declare(agentGrpc.CapabilityBasicAuthFirstColon, agentGrpc.CapabilityBasicAuthFile)
	f.agent.mu.Lock()
	f.agent.refusal = grpcStatus.Error(codes.InvalidArgument, "invalid basic authentication file '/agent/users': line '2': expected <username:password>")
	f.agent.mu.Unlock()
	code, output := f.run(t, nil, "--basic-auth-file", "/agent/users")
	require.Equal(t, exitPermanent, code, output)
	require.Contains(t, output, "line '2'")
	require.EqualValues(t, 1, f.agent.shares.Load())
}
