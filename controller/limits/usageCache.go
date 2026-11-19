package limits

import (
	"time"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/metrics"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
)

// usageCacheSize bounds each of the fast path's maps. the number of active accounts is far below it; the bound
// exists so the maps cannot grow without limit.
const usageCacheSize = 100000

// usageCaches lets the run loop finish an event whose conclusion is already known: the account is limitless,
// nothing applying to the event's backend mode is exceeded, or the exceeded class's crossing is already
// journaled. it may only ever conclude that there is nothing to do; anything that journals or acts is decided
// by enforce from the store and InfluxDB. it belongs to the agent's run goroutine and takes no lock.
type usageCaches struct {
	now        func() time.Time
	size       int
	accounts   map[int64]*accountEntry
	totals     map[totalsKey]totalsEntry
	lastReport time.Time
	finished   int64
	full       int64
}

type accountEntry struct {
	limitless bool
	ul        *userLimits
	expires   time.Time
	// crossings holds, per class, when the journal last confirmed a crossing of that class. each record is
	// trusted for one account lifetime from its own confirmation, and refreshing the account does not extend
	// it: otherwise events for one backend mode would keep alive a stale record for another mode's class.
	crossings map[crossingKey]time.Time
}

// crossingKey identifies a bandwidth class the way isBandwidthClassLimitedForAccount tells journal entries
// apart: a global class by its limit action, which separates the global warning from the global limit, and
// an applied limit class by its limit class id.
type crossingKey struct {
	global       bool
	action       store.LimitAction
	limitClassId int
}

func crossingKeyFor(bwc store.BandwidthClass) crossingKey {
	if bwc.IsGlobal() {
		return crossingKey{global: true, action: bwc.GetLimitAction()}
	}
	return crossingKey{limitClassId: bwc.GetLimitClassId()}
}

type totalsKey struct {
	accountId     int64
	periodMinutes int
}

type totalsEntry struct {
	rx      int64
	tx      int64
	expires time.Time
}

// enforcement is what one successful run of enforce read from the store and InfluxDB. it flows out of
// enforce to refresh the caches; nothing flows from the caches into enforce.
type enforcement struct {
	limitless bool
	ul        *userLimits
	totals    map[int]periodBwValues
	// confirmed is the class found exceeded whose journal entry existed or was just created.
	confirmed store.BandwidthClass
}

func (c *usageCaches) clock() time.Time {
	if c.now == nil {
		return time.Now()
	}
	return c.now()
}

func (c *usageCaches) ensure() {
	if c.accounts == nil {
		c.accounts = make(map[int64]*accountEntry)
	}
	if c.totals == nil {
		c.totals = make(map[totalsKey]totalsEntry)
	}
	if c.size <= 0 {
		c.size = usageCacheSize
	}
}

// handleUsage finishes the event from the caches when they show there is nothing to do, and otherwise runs
// enforce and refreshes the caches from what it read.
func (a *Agent) handleUsage(u *metrics.Usage) error {
	a.caches.ensure()
	if a.finishedFromCache(u) {
		a.caches.finished++
		return nil
	}
	a.caches.full++
	read, err := a.enforceSafely(u)
	if err != nil {
		return err
	}
	a.rememberEnforcement(u.AccountId, read)
	return nil
}

// finishedFromCache reads the maps and nothing else. it opens no transaction, calls no store or bandwidth
// reader method, and changes nothing outside the caches.
func (a *Agent) finishedFromCache(u *metrics.Usage) bool {
	// without an account or a backend mode the fast path cannot know which classes apply.
	if u.AccountId == 0 || u.BackendMode == "" {
		return false
	}
	lifetime := a.cfg.accountCacheLifetime()
	if lifetime < 0 {
		return false
	}
	now := a.caches.clock()
	entry, found := a.caches.accounts[u.AccountId]
	if !found || !now.Before(entry.expires) {
		return false
	}
	if entry.limitless {
		return true
	}
	// judge the class enforce would select: the last exceeded class in this backend mode's array.
	var selected store.BandwidthClass
	for _, bwc := range entry.ul.toBandwidthArray(sdk.BackendMode(u.BackendMode)) {
		totals, found := a.caches.totals[totalsKey{u.AccountId, bwc.GetPeriodMinutes()}]
		if !found || !now.Before(totals.expires) {
			return false
		}
		if a.transferBytesExceeded(totals.rx, totals.tx, bwc) {
			selected = bwc
		}
	}
	if selected == nil {
		return true
	}
	confirmed, found := entry.crossings[crossingKeyFor(selected)]
	return found && now.Before(confirmed.Add(lifetime))
}

func (a *Agent) rememberEnforcement(acctId int64, read *enforcement) {
	if acctId == 0 || read == nil {
		return
	}
	c := &a.caches
	now := c.clock()

	if lifetime := a.cfg.accountCacheLifetime(); lifetime >= 0 {
		entry, found := c.accounts[acctId]
		if !found {
			makeRoom(c.accounts, c.size, func(e *accountEntry) bool { return !now.Before(e.expires) })
			entry = &accountEntry{crossings: make(map[crossingKey]time.Time)}
			c.accounts[acctId] = entry
		}
		entry.limitless = read.limitless
		entry.ul = read.ul
		entry.expires = now.Add(lifetime)
		for key, confirmed := range entry.crossings {
			if !now.Before(confirmed.Add(lifetime)) {
				delete(entry.crossings, key)
			}
		}
		if read.confirmed != nil {
			entry.crossings[crossingKeyFor(read.confirmed)] = now
		}
	}

	if lifetime := a.cfg.usageCacheLifetime(); lifetime >= 0 {
		for period, totals := range read.totals {
			key := totalsKey{acctId, period}
			if _, found := c.totals[key]; !found {
				makeRoom(c.totals, c.size, func(e totalsEntry) bool { return !now.Before(e.expires) })
			}
			c.totals[key] = totalsEntry{rx: totals.rx, tx: totals.tx, expires: now.Add(lifetime)}
		}
	}
}

// makeRoom keeps a full map under its bound before an insert: it drops expired entries, and clears the map if
// none had expired.
func makeRoom[K comparable, V any](m map[K]V, size int, expired func(V) bool) {
	if len(m) < size {
		return
	}
	for k, v := range m {
		if expired(v) {
			delete(m, k)
		}
	}
	if len(m) >= size {
		clear(m)
	}
}

// reportCaches logs the fast-path counts once the report interval has passed, when events were handled
// since the last report. it is called from the run loop after each event.
func (a *Agent) reportCaches() {
	interval := a.cfg.cacheReportInterval()
	if interval < 0 {
		return
	}
	c := &a.caches
	now := c.clock()
	if now.Sub(c.lastReport) < interval {
		return
	}
	c.lastReport = now
	if c.finished+c.full == 0 {
		return
	}
	dl.Infof("limits fast path finished '%d' events, '%d' events took the full path", c.finished, c.full)
	c.finished, c.full = 0, 0
}
