package metrics

import (
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnusableUsageNacksWithoutRunningSinks(t *testing.T) {
	for _, tc := range []struct {
		name string
		old  string
		new  string
	}{
		{"wrong namespace", `"fabric.usage"`, `"other"`},
		{"missing namespace", `"namespace":"fabric.usage",`, ``},
		{"mistyped namespace", `"namespace":"fabric.usage"`, `"namespace":17`},
		{"missing interval", `"interval_start_utc":1,`, ``},
		{"mistyped interval", `"interval_start_utc":1`, `"interval_start_utc":"1"`},
		{"missing service", `"serviceId":"unrelated"`, ``},
		{"mistyped service", `"serviceId":"unrelated"`, `"serviceId":1`},
		{"missing tags", `"tags":{"serviceId":"unrelated"},`, ``},
		{"mistyped tags", `"tags":{"serviceId":"unrelated"}`, `"tags":[]`},
		{"missing usage", `,"usage":{}`, ``},
		{"mistyped usage", `"usage":{}`, `"usage":[]`},
		{"null usage", `"usage":{}`, `"usage":null`},
		{"mistyped ingress tx", `"usage":{}`, `"usage":{"ingress.tx":"1"}`},
		{"mistyped ingress rx", `"usage":{}`, `"usage":{"ingress.rx":"1"}`},
		{"mistyped egress tx", `"usage":{}`, `"usage":{"egress.tx":"1"}`},
		{"mistyped egress rx", `"usage":{}`, `"usage":{"egress.rx":"1"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := ZitiEventJson(strings.Replace(string(validUsage), tc.old, tc.new, 1))
			usage, err := Ingest(data)
			require.Error(t, err)
			require.Nil(t, usage)
			e := &testEvent{data: data, done: make(chan bool, 2)}
			var calls atomic.Int32
			startTestAgent(t, AgentConfig{}, []ZitiEventMsg{e}, sinkFunc(func(*Usage) error { calls.Add(1); return nil }))
			require.False(t, awaitEvent(t, e))
			require.EqualValues(t, 1, e.nacks.Load())
			require.Zero(t, e.acks.Load())
			require.Zero(t, calls.Load())
		})
	}
}

func TestUsageWithoutCircuitIsAccepted(t *testing.T) {
	usage, err := Ingest(validUsage)
	require.NoError(t, err)
	require.Empty(t, usage.ZitiCircuitId)
	require.Equal(t, "unrelated", usage.ZitiServiceId)
	e := &testEvent{data: validUsage, done: make(chan bool, 2)}
	var calls atomic.Int32
	startTestAgent(t, AgentConfig{}, []ZitiEventMsg{e}, sinkFunc(func(*Usage) error { calls.Add(1); return nil }))
	require.True(t, awaitEvent(t, e))
	require.EqualValues(t, 1, e.acks.Load())
	require.Zero(t, e.nacks.Load())
	require.EqualValues(t, 1, calls.Load())
}

func TestUsageWithoutEgressIsAccepted(t *testing.T) {
	data := ZitiEventJson(strings.Replace(string(validUsage), `"usage":{}`, `"usage":{"ingress.tx":123,"ingress.rx":456}`, 1))
	usage, err := Ingest(data)
	require.NoError(t, err)
	require.EqualValues(t, 123, usage.FrontendTx)
	require.EqualValues(t, 456, usage.FrontendRx)
	require.Zero(t, usage.BackendTx)
	require.Zero(t, usage.BackendRx)
	e := &testEvent{data: data, done: make(chan bool, 2)}
	var calls atomic.Int32
	startTestAgent(t, AgentConfig{}, []ZitiEventMsg{e}, sinkFunc(func(*Usage) error { calls.Add(1); return nil }))
	require.True(t, awaitEvent(t, e))
	require.EqualValues(t, 1, e.acks.Load())
	require.Zero(t, e.nacks.Load())
	require.EqualValues(t, 1, calls.Load())
}
