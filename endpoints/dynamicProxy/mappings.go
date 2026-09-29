package dynamicProxy

import (
	"context"
	"sync"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/michaelquigley/df/da"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/dynamicProxyController"
	"github.com/pkg/errors"
)

// mappingSource supplies frontend mappings; id 0 returns the complete set for the frontend, any other id returns
// only the rows with a higher id. the controller client satisfies it.
type mappingSource interface {
	getAllFrontendMappings(frontendToken string, id int64) ([]*dynamicProxyController.FrontendMapping, error)
}

type mappings struct {
	cfg          *config
	source       mappingSource
	updates      <-chan *dynamicProxyController.Mapping
	retryInitial time.Duration
	retryMax     time.Duration
	ctx          context.Context
	cancel       context.CancelFunc
	mutex        sync.RWMutex
	nameMap      map[string]*dynamicProxyController.FrontendMapping
	emptyPulls   int // consecutive empty reconciliation pulls held back; see applyReconcile
}

func buildMappings(app *da.Application[*config]) error {
	if app.Cfg.MappingReconcileInterval <= 0 {
		return errors.Errorf("mapping_reconcile_interval must be positive, got '%v'", app.Cfg.MappingReconcileInterval)
	}
	mappings := newMappings()
	mappings.cfg = app.Cfg
	da.Set(app.C, mappings)
	return nil
}

func newMappings() *mappings {
	return &mappings{
		retryInitial: time.Second,
		retryMax:     30 * time.Second,
		nameMap:      make(map[string]*dynamicProxyController.FrontendMapping),
	}
}

func (m *mappings) getMapping(name string) (*dynamicProxyController.FrontendMapping, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	mapping, exists := m.nameMap[name]
	return mapping, exists
}

func (m *mappings) Link(c *da.Container) error {
	subscriber, found := da.Get[*amqpSubscriber](c)
	if !found {
		return errors.New("no amqp subscriber found")
	}
	m.updates = subscriber.Updates()

	ctrl, found := da.Get[*controllerClient](c)
	if !found {
		return errors.New("no controller client found")
	}
	m.source = ctrl
	return nil
}

func (m *mappings) Start() error {
	m.ctx, m.cancel = context.WithCancel(context.Background())
	go m.run()
	return nil
}

func (m *mappings) Stop() error {
	if m.cancel != nil {
		m.cancel()
	}
	return nil
}

func (m *mappings) run() {
	dl.Infof("started")
	defer dl.Infof("stopped")

	if !m.initialLoad() {
		return
	}

	// the delta refresh and the full reconciliation run on independent intervals
	refresh := time.NewTicker(m.cfg.MappingRefreshInterval)
	defer refresh.Stop()
	reconcile := time.NewTicker(m.cfg.MappingReconcileInterval)
	defer reconcile.Stop()

	for {
		dl.ChannelLog("mappings").Debugf("\n%s", m.dumpMappings())
		select {
		case <-m.ctx.Done():
			return

		case <-refresh.C:
			m.refresh()

		case <-reconcile.C:
			m.reconcile()

		case update := <-m.updates:
			// handle real-time mapping updates from AMQP
			m.handleMappingUpdate(update)
		}
	}
}

// initialLoad pulls the complete set, retrying with backoff until it succeeds or the context is cancelled. until it
// succeeds the map is empty and requests answer not-found. returns false when cancelled.
func (m *mappings) initialLoad() bool {
	backoff := m.retryInitial
	for {
		start := time.Now()
		mappings, err := m.source.getAllFrontendMappings(m.cfg.FrontendToken, 0)
		if err == nil {
			m.updateMappings(mappings)
			dl.Infof("retrieved '%d' mappings in '%v'", len(mappings), time.Since(start))
			return true
		}
		dl.Errorf("failed to retrieve initial mappings (retrying in '%v'): %v", backoff, err)
		select {
		case <-time.After(backoff):
		case <-m.ctx.Done():
			return false
		}
		backoff = min(backoff*2, m.retryMax)
	}
}

// refresh pulls every row with a higher id than any held and applies it. that includes a replacement, since a name
// deleted and re-inserted gets a higher id, but never a deletion: a deleted row never shows up in a higher-id pull, so
// removals are left to the amqp unbinds and to reconcile.
func (m *mappings) refresh() {
	start := time.Now()
	highestId := m.getHighestId()
	mappings, err := m.source.getAllFrontendMappings(m.cfg.FrontendToken, highestId)
	if err != nil {
		dl.Errorf("failed to refresh mappings (highest version '%v'): %v", highestId, err)
		return
	}
	if len(mappings) > 0 {
		m.updateMappings(mappings)
		dl.Warnf("refresh updated '%d' mappings (highest version '%v') in '%v'", len(mappings), highestId, time.Since(start))
	} else {
		dl.Debugf("refresh found no new mappings (highest version '%v') in '%v'", highestId, time.Since(start))
	}
}

// reconcile pulls the complete set and makes the map match it, which recovers the removals the delta refresh cannot
// see and the amqp subscriber may have lost. a failed pull leaves the map untouched.
func (m *mappings) reconcile() {
	start := time.Now()
	full, err := m.source.getAllFrontendMappings(m.cfg.FrontendToken, 0)
	if err != nil {
		dl.Errorf("failed to pull mappings for reconciliation: %v", err)
		m.mutex.Lock()
		m.emptyPulls = 0
		m.mutex.Unlock()
		return
	}
	res, applied := m.applyReconcile(full)
	if !applied {
		return
	}
	dl.Infof("reconciled '%d' mappings: '%d' added, '%d' replaced, '%d' dropped in '%v'",
		res.total, res.added, res.replaced, res.dropped, time.Since(start))
}

type reconcileResult struct {
	total    int
	added    int
	replaced int
	dropped  int
}

// applyReconcile diffs the complete set against the map under the lock and applies the differences. returns false
// when the set was not applied.
func (m *mappings) applyReconcile(full []*dynamicProxyController.FrontendMapping) (reconcileResult, bool) {
	want := make(map[string]*dynamicProxyController.FrontendMapping, len(full))
	for _, mapping := range full {
		if mapping.Name != "" {
			want[mapping.Name] = mapping
		}
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// an empty complete set over a non-empty map is held back the first time. for a frontend carrying any traffic a
	// single empty answer is more likely a failure the client did not see than a world with no shares, and applying it
	// would drop every mapping and black-hole the frontend. the very next reconciliation returning empty as well is
	// applied: two consecutive empty answers an interval apart are taken as the truth, so a frontend whose last
	// mappings really were removed does not keep serving them until a restart. a non-empty set or a failed pull in
	// between resets the count.
	if len(want) == 0 && len(m.nameMap) > 0 && m.emptyPulls == 0 {
		m.emptyPulls++
		dl.Warnf("reconciliation pulled an empty mapping set while holding '%d' mappings; not applied unless the next reconciliation is also empty", len(m.nameMap))
		return reconcileResult{}, false
	}
	m.emptyPulls = 0

	res := reconcileResult{total: len(want)}
	for name, held := range m.nameMap {
		if _, found := want[name]; !found {
			delete(m.nameMap, name)
			res.dropped++
			dl.Infof("reconcile dropped mapping '%v' ('%v' -> none)", name, held.ShareToken)
		}
	}
	for name, mapping := range want {
		held, found := m.nameMap[name]
		switch {
		case !found:
			m.nameMap[name] = mapping
			res.added++
			dl.Infof("reconcile added mapping '%v' (none -> '%v')", name, mapping.ShareToken)
		case held.ShareToken != mapping.ShareToken || held.Id != mapping.Id:
			m.nameMap[name] = mapping
			res.replaced++
			dl.Infof("reconcile replaced mapping '%v' ('%v' id '%d' -> '%v' id '%d')", name, held.ShareToken, held.Id, mapping.ShareToken, mapping.Id)
		}
	}
	return res, true
}

func (m *mappings) getHighestId() int64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var highestId int64
	for _, mapping := range m.nameMap {
		if version := mapping.GetId(); version > highestId {
			highestId = version
		}
	}
	return highestId
}

func (m *mappings) handleMappingUpdate(update *dynamicProxyController.Mapping) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	switch update.Operation {
	case dynamicProxyController.OperationBind:
		mapping := &dynamicProxyController.FrontendMapping{
			Id:         update.Id,
			Name:       update.Name,
			ShareToken: update.ShareToken,
		}
		m.nameMap[mapping.Name] = mapping
		dl.Infof("added mapping: '%v' -> '%v'", mapping.Name, mapping.ShareToken)

	case dynamicProxyController.OperationUnbind:
		delete(m.nameMap, update.Name)
		dl.Infof("removed mapping: '%v'", update.Name)

	default:
		dl.Errorf("unknown mapping operation '%v'", update.Operation)
	}
}

func (m *mappings) updateMappings(frontendMappings []*dynamicProxyController.FrontendMapping) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// populate new mappings
	for _, mapping := range frontendMappings {
		if mapping.Name != "" {
			m.nameMap[mapping.Name] = mapping
		}
		dl.Infof("added mapping: '%v' -> '%v'", mapping.Name, mapping.ShareToken)
	}
}

// dumpMappings returns a formatted table string containing all mapping details
func (m *mappings) dumpMappings() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	t := table.NewWriter()
	t.SetStyle(table.StyleRounded)
	t.SetCaption("%d mappings", len(m.nameMap))
	t.AppendHeader(table.Row{"name", "share token"})
	for key, mapping := range m.nameMap {
		t.AppendRow(table.Row{key, mapping.ShareToken})
	}
	return t.Render()
}
