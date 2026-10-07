package controller

import (
	"errors"
	"slices"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/v2/controller/agentController"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/agent"
	"github.com/stretchr/testify/require"
)

func (f *shareCreateFixture) enrollAgent(h *agentEnrollHandler) interface{} {
	cfg.AgentController = &agentController.Config{ZId: "agent-controller-zid"}
	return h.Handle(agent.EnrollParams{Body: agent.EnrollBody{EnvZID: "env-zid"}}, f.principal)
}

func (f *shareCreateFixture) agentEnrolled(t *testing.T) bool {
	t.Helper()
	envId := f.fixtureEnvironmentId(t)
	trx, err := str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()
	enrolled, err := str.IsAgentEnrolledForEnvironment(envId, trx)
	require.NoError(t, err)
	return enrolled
}

func TestAgentEnrollCommitFailureCompensates(t *testing.T) {
	f := setupShareCreateFixture(t)
	h := newAgentEnrollHandler()
	h.commit = func(*sqlx.Tx) error { return errors.New("commit failed") }

	resp := f.enrollAgent(h)

	require.IsType(t, &agent.EnrollInternalServerError{}, resp)
	created, deleted := f.fake.Log()
	require.Len(t, created, 4)
	slices.Reverse(created)
	require.Equal(t, created, deleted)
	require.False(t, f.agentEnrolled(t))
	require.Contains(t, f.logs.String(), "compensated failed agent enrollment")
}

func TestAgentEnrollSuccessKeepsObjects(t *testing.T) {
	f := setupShareCreateFixture(t)

	resp := f.enrollAgent(newAgentEnrollHandler())

	require.IsType(t, &agent.EnrollOK{}, resp)
	created, deleted := f.fake.Log()
	require.Len(t, created, 4)
	require.Empty(t, deleted)
	require.True(t, f.agentEnrolled(t))
}
