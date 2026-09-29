package dynamicProxy

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/dynamicProxyController"
	"github.com/stretchr/testify/require"
)

// fakeSource answers each pull from its queue of responses, repeating the last one when the queue runs out.
type fakeSource struct {
	mutex     sync.Mutex
	responses []fakeResponse
	calls     []int64
}

type fakeResponse struct {
	mappings []*dynamicProxyController.FrontendMapping
	err      error
}

func (f *fakeSource) getAllFrontendMappings(_ string, id int64) ([]*dynamicProxyController.FrontendMapping, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.calls = append(f.calls, id)
	r := f.responses[0]
	if len(f.responses) > 1 {
		f.responses = f.responses[1:]
	}
	return r.mappings, r.err
}

func (f *fakeSource) callCount() int {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return len(f.calls)
}

func fm(id int64, name, shareToken string) *dynamicProxyController.FrontendMapping {
	return &dynamicProxyController.FrontendMapping{Id: id, Name: name, ShareToken: shareToken}
}

func newTestMappings(source mappingSource, held ...*dynamicProxyController.FrontendMapping) *mappings {
	m := newMappings()
	m.cfg = &config{
		FrontendToken:            "fe",
		MappingRefreshInterval:   time.Hour,
		MappingReconcileInterval: time.Hour,
	}
	m.source = source
	for _, mapping := range held {
		m.nameMap[mapping.Name] = mapping
	}
	return m
}

func captureLogs(t *testing.T) *bytes.Buffer {
	var logs bytes.Buffer
	dl.Init(dl.DefaultOptions().JSON().SetOutput(&logs))
	t.Cleanup(func() { dl.Init() })
	return &logs
}

func requireMapping(t *testing.T, m *mappings, name, shareToken string, id int64) {
	t.Helper()
	mapping, found := m.getMapping(name)
	require.True(t, found, "mapping '%s' missing", name)
	require.Equal(t, shareToken, mapping.ShareToken)
	require.Equal(t, id, mapping.Id)
}

func TestReconcileDropsMissingName(t *testing.T) {
	source := &fakeSource{responses: []fakeResponse{{mappings: []*dynamicProxyController.FrontendMapping{fm(1, "a", "sa")}}}}
	m := newTestMappings(source, fm(1, "a", "sa"), fm(2, "b", "sb"))

	m.reconcile()

	require.Equal(t, []int64{0}, source.calls)
	_, found := m.getMapping("b")
	require.False(t, found)
	requireMapping(t, m, "a", "sa", 1)
}

func TestReconcileAddsMissingName(t *testing.T) {
	source := &fakeSource{responses: []fakeResponse{{mappings: []*dynamicProxyController.FrontendMapping{fm(1, "a", "sa"), fm(2, "b", "sb")}}}}
	m := newTestMappings(source, fm(1, "a", "sa"))

	m.reconcile()

	requireMapping(t, m, "a", "sa", 1)
	requireMapping(t, m, "b", "sb", 2)
}

func TestReconcileReplacesDifferingRow(t *testing.T) {
	same := fm(1, "a", "sa")
	res, applied := newTestMappings(nil).applyReconcile(nil)
	require.True(t, applied, "an empty set over an empty map is applied")
	require.Equal(t, reconcileResult{}, res)

	m := newTestMappings(nil, same, fm(2, "b", "sb-old"))
	res, applied = m.applyReconcile([]*dynamicProxyController.FrontendMapping{fm(1, "a", "sa"), fm(3, "b", "sb-new")})

	require.True(t, applied)
	require.Equal(t, reconcileResult{total: 2, replaced: 1}, res)
	requireMapping(t, m, "b", "sb-new", 3)
	held, _ := m.getMapping("a")
	require.Same(t, same, held, "an identical row is left alone")
}

func TestReconcileFailureLeavesMapUntouched(t *testing.T) {
	t.Run("failed pull", func(t *testing.T) {
		logs := captureLogs(t)
		source := &fakeSource{responses: []fakeResponse{{err: errors.New("unreachable")}}}
		m := newTestMappings(source, fm(1, "a", "sa"), fm(2, "b", "sb"))

		m.reconcile()

		requireMapping(t, m, "a", "sa", 1)
		requireMapping(t, m, "b", "sb", 2)
		require.Contains(t, logs.String(), "failed to pull mappings for reconciliation")
	})

	t.Run("empty set over non-empty map", func(t *testing.T) {
		logs := captureLogs(t)
		source := &fakeSource{responses: []fakeResponse{{}}}
		m := newTestMappings(source, fm(1, "a", "sa"), fm(2, "b", "sb"))

		m.reconcile()

		requireMapping(t, m, "a", "sa", 1)
		requireMapping(t, m, "b", "sb", 2)
		require.Contains(t, logs.String(), "reconciliation pulled an empty mapping set while holding '2' mappings; not applied")
		require.Contains(t, logs.String(), `"level":"WARN"`)

		// the very next reconciliation also empty is applied
		m.reconcile()

		require.Empty(t, m.nameMap)
		require.Contains(t, logs.String(), "reconcile dropped mapping 'a'")
		require.Contains(t, logs.String(), "reconcile dropped mapping 'b'")
		require.Contains(t, logs.String(), "reconciled '0' mappings: '0' added, '0' replaced, '2' dropped")
	})

	t.Run("non-empty set between empty sets", func(t *testing.T) {
		held := []*dynamicProxyController.FrontendMapping{fm(1, "a", "sa")}
		source := &fakeSource{responses: []fakeResponse{{}, {mappings: held}, {}}}
		m := newTestMappings(source, held...)

		m.reconcile()
		m.reconcile()
		m.reconcile()

		require.Equal(t, 3, source.callCount())
		requireMapping(t, m, "a", "sa", 1)
	})

	t.Run("failed pull between empty sets", func(t *testing.T) {
		source := &fakeSource{responses: []fakeResponse{{}, {err: errors.New("unreachable")}, {}}}
		m := newTestMappings(source, fm(1, "a", "sa"))

		m.reconcile()
		m.reconcile()
		m.reconcile()

		require.Equal(t, 3, source.callCount())
		requireMapping(t, m, "a", "sa", 1)
	})
}

func TestRefreshOnlyAdds(t *testing.T) {
	// the delta pull answers with a new row only; the rows it does not mention must survive
	source := &fakeSource{responses: []fakeResponse{{mappings: []*dynamicProxyController.FrontendMapping{fm(3, "c", "sc")}}}}
	m := newTestMappings(source, fm(1, "a", "sa"), fm(2, "b", "sb"))

	m.refresh()

	require.Equal(t, []int64{2}, source.calls, "refresh asks for rows above the highest id held")
	requireMapping(t, m, "a", "sa", 1)
	requireMapping(t, m, "b", "sb", 2)
	requireMapping(t, m, "c", "sc", 3)

	// an empty delta drops nothing
	source.responses = []fakeResponse{{}}
	m.refresh()
	require.Len(t, m.nameMap, 3)
}

func TestStartupRetriesUntilLoaded(t *testing.T) {
	source := &fakeSource{responses: []fakeResponse{
		{err: errors.New("unreachable")},
		{err: errors.New("unreachable")},
		{mappings: []*dynamicProxyController.FrontendMapping{fm(1, "a", "sa")}},
	}}
	updates := make(chan *dynamicProxyController.Mapping, 1)
	m := newTestMappings(source)
	m.updates = updates
	m.retryInitial = time.Millisecond
	m.retryMax = 2 * time.Millisecond

	require.NoError(t, m.Start())
	t.Cleanup(func() { _ = m.Stop() })

	require.Eventually(t, func() bool {
		_, found := m.getMapping("a")
		return found
	}, 5*time.Second, time.Millisecond)
	require.Equal(t, 3, source.callCount())

	// the loop is still running after the retries: it applies a live update
	updates <- &dynamicProxyController.Mapping{Id: 2, Operation: dynamicProxyController.OperationBind, Name: "b", ShareToken: "sb"}
	require.Eventually(t, func() bool {
		_, found := m.getMapping("b")
		return found
	}, 5*time.Second, time.Millisecond)
}

func TestStartupRetryStopsOnCancel(t *testing.T) {
	source := &fakeSource{responses: []fakeResponse{{err: errors.New("unreachable")}}}
	m := newTestMappings(source)
	m.retryInitial = time.Hour
	m.ctx, m.cancel = context.WithCancel(context.Background())
	m.cancel()

	require.False(t, m.initialLoad())
	require.Equal(t, 1, source.callCount())
}
