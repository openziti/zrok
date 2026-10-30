package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
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
