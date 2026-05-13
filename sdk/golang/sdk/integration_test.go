package sdk

// Integration tests for the zrok2 Go SDK against a live instance.
//
// Required environment variables:
//
//	ZROK2_API_ENDPOINT  — controller URL (e.g. http://localhost:18080)
//	ZROK2_ADMIN_TOKEN   — admin secret for account creation
//
// Run:
//
//	ZROK2_API_ENDPOINT=http://localhost:18080 \
//	ZROK2_ADMIN_TOKEN=<token> \
//	go test ./sdk/golang/sdk/... -run Integration -v

import (
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"testing"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/runtime"
	"github.com/openziti/zrok/v2/environment/env_core"
	"github.com/openziti/zrok/v2/environment/env_v0_4"
	"github.com/openziti/zrok/v2/rest_client_zrok"
	"github.com/openziti/zrok/v2/rest_client_zrok/admin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// integrationEnv holds the fixtures for a single integration test run.
type integrationEnv struct {
	apiEndpoint string
	adminToken  string
	root        env_core.Root
}

// setupIntegration skips the test if env vars are missing, creates a
// temporary account and enabled environment, and returns the fixtures.
// The returned cleanup function must be deferred by the caller.
func setupIntegration(t *testing.T) (*integrationEnv, func()) {
	t.Helper()

	apiEndpoint := os.Getenv("ZROK2_API_ENDPOINT")
	if apiEndpoint == "" {
		t.Skip("ZROK2_API_ENDPOINT not set — skipping integration tests")
	}
	adminToken := os.Getenv("ZROK2_ADMIN_TOKEN")
	if adminToken == "" {
		t.Skip("ZROK2_ADMIN_TOKEN not set — skipping integration tests")
	}

	// Redirect the SDK's root dir to a per-test temp directory so tests are
	// fully isolated from the operator's ~/.zrok2 and from each other.
	tmpDir := t.TempDir()
	env_v0_4.SetRootDirName(tmpDir)

	// Create a unique test account via the admin API.
	apiURL, err := url.Parse(apiEndpoint)
	require.NoError(t, err, "parse API endpoint")
	transport := httptransport.New(apiURL.Host, "/api/v2", []string{apiURL.Scheme})
	transport.Producers["application/zrok.v1+json"] = runtime.JSONProducer()
	transport.Consumers["application/zrok.v1+json"] = runtime.JSONConsumer()
	zrokClient := rest_client_zrok.New(transport, nil)
	adminAuth := httptransport.APIKeyAuth("X-TOKEN", "header", adminToken)

	email := fmt.Sprintf("integ-%s-%08x@zrok.internal", t.Name(), rand.Uint32())
	password := "integration-test-password-1234"
	createReq := admin.NewCreateAccountParams()
	createReq.Body = admin.CreateAccountBody{
		Email:    email,
		Password: password,
	}
	createResp, err := zrokClient.Admin.CreateAccount(createReq, adminAuth)
	require.NoError(t, err, "create test account")
	accountToken := createResp.Payload.AccountToken

	// Build a Root pointed at the local controller with the new account token.
	root, err := env_v0_4.Default()
	require.NoError(t, err, "create root")
	err = root.SetConfig(&env_core.Config{ApiEndpoint: apiEndpoint})
	require.NoError(t, err, "set config")
	err = root.SetEnvironment(&env_core.Environment{AccountToken: accountToken, ApiEndpoint: apiEndpoint})
	require.NoError(t, err, "set environment (pre-enable)")

	// Enable the environment (creates Ziti identity, stores it in tmpDir).
	env, err := EnableEnvironment(root, &EnableRequest{
		Description: fmt.Sprintf("integration-test/%s", t.Name()),
	})
	require.NoError(t, err, "enable environment")
	err = root.SetEnvironment(&env_core.Environment{
		AccountToken: accountToken,
		ZitiIdentity: env.ZitiIdentity,
		ApiEndpoint:  apiEndpoint,
	})
	require.NoError(t, err, "store enabled environment")

	cleanup := func() {
		if root.IsEnabled() {
			_ = DisableEnvironment(env, root)
		}
		// tmpDir is cleaned by t.TempDir()
	}

	return &integrationEnv{
		apiEndpoint: apiEndpoint,
		adminToken:  adminToken,
		root:        root,
	}, cleanup
}

// TestIntegrationEnableDisable verifies that EnableEnvironment and
// DisableEnvironment work correctly against a live controller.
func TestIntegrationEnableDisable(t *testing.T) {
	env, cleanup := setupIntegration(t)
	defer cleanup()

	assert.True(t, env.root.IsEnabled(), "environment should be enabled after setup")
	assert.NotEmpty(t, env.root.Environment().ZitiIdentity, "ZitiIdentity should be set")
}

// TestIntegrationShareLifecycle creates a private share, verifies detail,
// then deletes it.
func TestIntegrationShareLifecycle(t *testing.T) {
	ie, cleanup := setupIntegration(t)
	defer cleanup()

	shr, err := CreateShare(ie.root, &ShareRequest{
		ShareMode:   PrivateShareMode,
		BackendMode: TcpTunnelBackendMode,
		Target:      "tcp://localhost:9999",
	})
	require.NoError(t, err, "create share")
	require.NotEmpty(t, shr.Token, "share token should not be empty")

	// Verify detail is retrievable.
	detail, err := GetShareDetail(ie.root, shr.Token)
	require.NoError(t, err, "get share detail")
	assert.Equal(t, shr.Token, detail.ShareToken, "share token should match")

	// Clean up.
	err = DeleteShare(ie.root, shr)
	assert.NoError(t, err, "delete share")
}

// TestIntegrationAccessLifecycle creates a private share, creates an access
// for it, then tears both down.
func TestIntegrationAccessLifecycle(t *testing.T) {
	ie, cleanup := setupIntegration(t)
	defer cleanup()

	shr, err := CreateShare(ie.root, &ShareRequest{
		ShareMode:   PrivateShareMode,
		BackendMode: TcpTunnelBackendMode,
		Target:      "tcp://localhost:9998",
	})
	require.NoError(t, err, "create share")
	defer func() { _ = DeleteShare(ie.root, shr) }()

	acc, err := CreateAccess(ie.root, &AccessRequest{
		ShareToken: shr.Token,
	})
	require.NoError(t, err, "create access")
	require.NotEmpty(t, acc.Token, "access token should not be empty")
	assert.Equal(t, shr.Token, acc.ShareToken, "access share token should match")

	err = DeleteAccess(ie.root, acc)
	assert.NoError(t, err, "delete access")
}

// TestIntegrationOverview verifies Overview returns a non-empty JSON response.
func TestIntegrationOverview(t *testing.T) {
	ie, cleanup := setupIntegration(t)
	defer cleanup()

	out, err := Overview(ie.root)
	require.NoError(t, err, "overview")
	assert.NotEmpty(t, out, "overview response should not be empty")
}
