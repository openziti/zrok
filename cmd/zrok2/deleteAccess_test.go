package main

import (
	"testing"

	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveAccessDeleteEnvZIdUsesOverrideWhenPresent(t *testing.T) {
	envZId := resolveAccessDeleteEnvZId("current-env", "override-env")
	assert.Equal(t, "override-env", envZId)
}

func TestResolveAccessDeleteEnvZIdUsesCurrentWhenOverrideEmpty(t *testing.T) {
	envZId := resolveAccessDeleteEnvZId("current-env", "")
	assert.Equal(t, "current-env", envZId)
}

func TestResolveUnaccessRequestMatchesFrontendToken(t *testing.T) {
	req, err := resolveUnaccessRequest("fe-123", "env-123", []*rest_model_zrok.AccessSummary{
		{FrontendToken: "fe-other", ShareToken: "shr-other"},
		{FrontendToken: "fe-123", ShareToken: "shr-123"},
	})

	require.NoError(t, err)
	require.NotNil(t, req)
	assert.Equal(t, "fe-123", req.Body.FrontendToken)
	assert.Equal(t, "shr-123", req.Body.ShareToken)
	assert.Equal(t, "env-123", req.Body.EnvZID)
}

func TestResolveUnaccessRequestReturnsNotFoundWhenFrontendMissing(t *testing.T) {
	req, err := resolveUnaccessRequest("fe-missing", "env-123", []*rest_model_zrok.AccessSummary{
		{FrontendToken: "fe-123", ShareToken: "shr-123"},
	})

	require.Error(t, err)
	assert.Nil(t, req)
	assert.Contains(t, err.Error(), "access 'fe-missing' not found")
}

func TestResolveUnaccessRequestRequiresShareToken(t *testing.T) {
	req, err := resolveUnaccessRequest("fe-123", "env-123", []*rest_model_zrok.AccessSummary{
		{FrontendToken: "fe-123"},
	})

	require.Error(t, err)
	assert.Nil(t, req)
	assert.Contains(t, err.Error(), "has no associated share token")
}
