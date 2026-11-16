package agent

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/agent/agentGrpc"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

// basicAuthFileSentinel is a distinctive password; the "never" fixtures assert it appears nowhere it could be read
// later.
const basicAuthFileSentinel = "zrokSentinelPw7Q"

// referenceCreates guards and records every controller create for a lifecycle agent whose public requests are all
// by reference. a create whose prepared request carries no users fails the test: that would be a share created
// without the authentication that was asked for.
type referenceCreates struct {
	mu       sync.Mutex
	prepared [][]string
}

func guardReferenceCreates(t *testing.T, a *Agent) *referenceCreates {
	creates := &referenceCreates{}
	a.deps.createPublic = func(_ context.Context, req *sdk.ShareRequest) (*sdk.Share, error) {
		if len(req.BasicAuth) == 0 {
			t.Errorf("created a by-reference share with no users")
		}
		creates.mu.Lock()
		creates.prepared = append(creates.prepared, slices.Clone(req.BasicAuth))
		creates.mu.Unlock()
		return &sdk.Share{Token: "known-token", FrontendEndpoints: []string{"http://share.example"}}, nil
	}
	return creates
}

func (c *referenceCreates) all() [][]string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return slices.Clone(c.prepared)
}

// lockedBuffer is a log destination safe for the agent's goroutines.
type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// captureAgentLogs sends all dl and logrus output, at every level, to a buffer for the rest of the test.
func captureAgentLogs(t *testing.T) *lockedBuffer {
	logs := &lockedBuffer{}
	opts := dl.DefaultOptions()
	opts.Output = logs
	opts.Level = slog.LevelDebug
	dl.Init(opts)
	previousOut, previousLevel := logrus.StandardLogger().Out, logrus.GetLevel()
	logrus.SetOutput(logs)
	logrus.SetLevel(logrus.TraceLevel)
	t.Cleanup(func() {
		dl.Init()
		logrus.SetOutput(previousOut)
		logrus.SetLevel(previousLevel)
	})
	return logs
}

func writeUsers(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))
}

func namedReferenceRequest(path, name string) *agentGrpc.SharePublicRequest {
	return &agentGrpc.SharePublicRequest{BackendMode: "proxy", Target: "8080", BasicAuthFile: path,
		NameSelections: []*agentGrpc.NameSelection{{NamespaceToken: "public", Name: name}}}
}

func operationFailure(t *testing.T, a *Agent) *FailureEntry {
	t.Helper()
	op := onlyOperation(t, a)
	a.mu.Lock()
	defer a.mu.Unlock()
	if op.failure == nil {
		return nil
	}
	copy := *op.failure
	return &copy
}

func TestBasicAuthFileMissingAtOriginalRequestCreatesNothing(t *testing.T) {
	a := newLifecycleAgent(t)
	creates := guardReferenceCreates(t, a)
	missing := filepath.Join(t.TempDir(), "users")
	_, err := publicErrorClient(t, a).SharePublic(context.Background(), namedReferenceRequest(missing, "app"))
	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	require.Contains(t, status.Convert(err).Message(), missing)
	require.Empty(t, creates.all())
	// a new request that fails before success gains no restart intent.
	require.NoError(t, a.Shutdown())
	require.Empty(t, savedRegistry(t, a).PublicShares)
}

func TestBasicAuthFileMissingAtReplayIsAFailedStartRetriedUntilItAppears(t *testing.T) {
	a := newLifecycleAgent(t)
	creates := guardReferenceCreates(t, a)
	users := filepath.Join(t.TempDir(), "users")
	path, err := a.root.AgentRegistry()
	require.NoError(t, err)
	saved := &SharePublicRequest{BackendMode: "proxy", Target: "8080", BasicAuthFile: users, NameSelections: []NameSelection{{NamespaceToken: "public", Name: "app"}}}
	require.NoError(t, (&Registry{PublicShares: []*PublicShareRegistryEntry{{Request: saved}}}).Save(path))

	require.NoError(t, a.RestoreRegistryOnce())
	require.Empty(t, creates.all())
	failure := operationFailure(t, a)
	require.NotNil(t, failure)
	require.Contains(t, failure.LastError, users)
	statusResponse, err := (&agentGrpcImpl{agent: a}).Status(context.Background(), &agentGrpc.StatusRequest{})
	require.NoError(t, err)
	require.Len(t, statusResponse.Shares, 1)
	require.Equal(t, "retrying", statusResponse.Shares[0].Status)
	require.NotNil(t, statusResponse.Shares[0].Failure)
	require.Contains(t, statusResponse.Shares[0].Failure.LastError, users, "status shows the failure with the path")
	registry := savedRegistry(t, a)
	require.Len(t, registry.PublicShares, 1, "the request keeps its restart intent")
	require.Equal(t, users, registry.PublicShares[0].Request.BasicAuthFile)

	// a retry while the file is still missing fails the same way.
	a.retryDue(time.Now().Add(time.Hour))
	require.Eventually(t, func() bool { return operationFailure(t, a).Count >= 2 }, 10*time.Second, 10*time.Millisecond)
	require.Empty(t, creates.all())

	// the file appears late, as a slow secret mount would; the next retry brings the share up with its users.
	writeUsers(t, users, "admin:s3:cret\n")
	a.retryDue(time.Now().Add(time.Hour))
	require.Eventually(t, func() bool { return len(creates.all()) == 1 }, 10*time.Second, 10*time.Millisecond)
	require.Equal(t, [][]string{{"admin:s3:cret"}}, creates.all())
}

func TestBasicAuthFileIsReadAtEachCreate(t *testing.T) {
	dir, err := os.MkdirTemp("", "agent-")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, os.RemoveAll(dir)) })
	users := filepath.Join(t.TempDir(), "users")
	writeUsers(t, users, "admin:first\n")

	first := newLifecycleAgentIn(t, dir)
	firstCreates := guardReferenceCreates(t, first)
	_, err = (&agentGrpcImpl{agent: first}).SharePublic(context.Background(), namedReferenceRequest(users, "app"))
	require.NoError(t, err)
	require.Equal(t, [][]string{{"admin:first"}}, firstCreates.all())
	require.NoError(t, first.Shutdown())

	t.Run("a changed file takes effect at replay", func(t *testing.T) {
		writeUsers(t, users, "admin:second\r\nother:third\r\n")
		second := newLifecycleAgentIn(t, dir)
		secondCreates := guardReferenceCreates(t, second)
		require.NoError(t, second.RestoreRegistryOnce())
		require.Equal(t, [][]string{{"admin:second", "other:third"}}, secondCreates.all())
		require.NoError(t, second.Shutdown())
	})

	t.Run("an emptied file creates nothing at replay", func(t *testing.T) {
		writeUsers(t, users, "\n")
		third := newLifecycleAgentIn(t, dir)
		thirdCreates := guardReferenceCreates(t, third)
		require.NoError(t, third.RestoreRegistryOnce())
		require.Empty(t, thirdCreates.all())
		failure := operationFailure(t, third)
		require.NotNil(t, failure)
		require.Contains(t, failure.LastError, users)
		require.NoError(t, third.Shutdown())
	})
}

func TestBasicAuthFileCredentialsAreNeverRetained(t *testing.T) {
	logs := captureAgentLogs(t)
	a := newLifecycleAgent(t)
	creates := guardReferenceCreates(t, a)
	users := filepath.Join(t.TempDir(), "users")
	writeUsers(t, users, "admin:"+basicAuthFileSentinel+"\n")
	client := publicErrorClient(t, a)

	_, err := client.SharePublic(context.Background(), namedReferenceRequest(users, "app"))
	require.NoError(t, err)
	require.Equal(t, [][]string{{"admin:" + basicAuthFileSentinel}}, creates.all(), "the users reach the create")

	op := onlyOperation(t, a)
	a.mu.Lock()
	request := cloneRequest(op.request).(*SharePublicRequest)
	a.mu.Unlock()
	require.Empty(t, request.BasicAuth, "the saved request holds the reference, never the users")
	require.Equal(t, users, request.BasicAuthFile)

	statusResponse, err := client.Status(context.Background(), &agentGrpc.StatusRequest{})
	require.NoError(t, err)
	statusJSON, err := protojson.Marshal(statusResponse)
	require.NoError(t, err)
	require.NotContains(t, string(statusJSON), basicAuthFileSentinel)

	require.NoError(t, a.SaveRegistry())
	path, err := a.root.AgentRegistry()
	require.NoError(t, err)
	registryBytes, err := os.ReadFile(path)
	require.NoError(t, err)
	require.NotContains(t, string(registryBytes), basicAuthFileSentinel)
	require.Contains(t, string(registryBytes), `"basic_auth_file"`)

	// a later malformed line: the refusal names the path and line, never the line.
	writeUsers(t, users, "admin:one\n"+basicAuthFileSentinel+"\n")
	_, err = client.SharePublic(context.Background(), namedReferenceRequest(users, "other"))
	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	require.Contains(t, err.Error(), "line '2'")
	require.NotContains(t, err.Error(), basicAuthFileSentinel)

	statusResponse, err = client.Status(context.Background(), &agentGrpc.StatusRequest{})
	require.NoError(t, err)
	statusJSON, err = protojson.Marshal(statusResponse)
	require.NoError(t, err)
	require.NotContains(t, string(statusJSON), basicAuthFileSentinel)

	require.NoError(t, a.Shutdown())
	registryBytes, err = os.ReadFile(path)
	require.NoError(t, err)
	require.NotContains(t, string(registryBytes), basicAuthFileSentinel)
	require.NotEmpty(t, logs.String(), "the log capture must see the agent's output")
	require.NotContains(t, logs.String(), basicAuthFileSentinel)
}

func TestBasicAuthFileDirectRequestConflictsAreRefused(t *testing.T) {
	users := filepath.Join(t.TempDir(), "users")
	writeUsers(t, users, "admin:secret\n")
	for name, request := range map[string]*agentGrpc.SharePublicRequest{
		"with users":          {BackendMode: "proxy", Target: "8080", BasicAuthFile: users, BasicAuth: []string{"admin:secret"}},
		"with oauth provider": {BackendMode: "proxy", Target: "8080", BasicAuthFile: users, OauthProvider: "github"},
	} {
		t.Run(name, func(t *testing.T) {
			a := newLifecycleAgent(t)
			creates := guardReferenceCreates(t, a)
			_, err := publicErrorClient(t, a).SharePublic(context.Background(), request)
			require.Error(t, err)
			require.Equal(t, codes.InvalidArgument, status.Code(err))
			require.Empty(t, creates.all())
		})
	}
}

func TestBasicAuthFileRelativePathResolvesInAgentEnvironment(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeUsers(t, filepath.Join(dir, "users"), "admin:secret\n")
	a := newLifecycleAgent(t)
	creates := guardReferenceCreates(t, a)
	_, err := (&agentGrpcImpl{agent: a}).SharePublic(context.Background(), &agentGrpc.SharePublicRequest{BackendMode: "proxy", Target: "8080", BasicAuthFile: "users"})
	require.NoError(t, err)
	require.Equal(t, [][]string{{"admin:secret"}}, creates.all())
}

func TestBasicAuthFileRefusalClassification(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "users")
	a := newLifecycleAgent(t)
	_, err := a.preparePublicRequest(&SharePublicRequest{BackendMode: "proxy", Target: "8080", BasicAuthFile: missing})
	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument, status.Code(publicShareRPCError(fmt.Errorf("wrapped: %w", err))))
	require.Equal(t, codes.InvalidArgument, status.Code(publicShareRPCError(errors.Join(err, errors.New("cleanup")))))
	require.Equal(t, codes.InvalidArgument, status.Code(publicShareRPCError(errBasicAuthFileConflict)))
	// other preparation errors keep their classification.
	_, err = a.preparePublicRequest(&SharePublicRequest{BackendMode: "proxy", Target: "8080", OauthRefreshInterval: "soon"})
	require.Error(t, err)
	require.NotEqual(t, codes.InvalidArgument, status.Code(publicShareRPCError(err)))
}

func TestBasicAuthFileRequestIdentityIsThePath(t *testing.T) {
	left := &SharePublicRequest{BackendMode: "proxy", Target: "8080", BasicAuthFile: "/etc/zrok2/users"}
	require.True(t, sameRequest(left, cloneRequest(left)))
	require.True(t, sameRequest(left, &SharePublicRequest{BackendMode: "proxy", Target: "8080", BasicAuthFile: "/etc/zrok2/users"}))
	require.False(t, sameRequest(left, &SharePublicRequest{BackendMode: "proxy", Target: "8080", BasicAuthFile: "/etc/zrok2/other"}))
	require.False(t, sameRequest(left, &SharePublicRequest{BackendMode: "proxy", Target: "8080"}))
}
