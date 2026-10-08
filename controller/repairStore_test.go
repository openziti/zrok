package controller

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/stretchr/testify/require"
)

type repairStoreFixture struct {
	logs *lockedBuffer
	acct int
	env  int
	ns   int
}

// setupRepairStoreFixture opens an in-memory store with one account, environment and namespace; the sweep
// needs nothing else.
func setupRepairStoreFixture(t *testing.T) *repairStoreFixture {
	t.Helper()
	prevStore := str
	t.Cleanup(func() { str = prevStore })
	logs := &lockedBuffer{}
	dl.Init(dl.DefaultOptions().JSON().SetOutput(logs))
	t.Cleanup(func() { dl.Init() })

	var err error
	str, err = store.Open(&store.Config{Path: ":memory:", Type: "sqlite3"})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, str.Close()) })

	f := &repairStoreFixture{logs: logs}
	trx, err := str.Begin()
	require.NoError(t, err)
	f.acct, err = str.CreateAccount(&store.Account{Email: "owner@example.com", Salt: "salt", Password: "password", Token: "acct-token"}, trx)
	require.NoError(t, err)
	f.env, err = str.CreateEnvironment(f.acct, &store.Environment{Description: "env", Host: "host", Address: "address", ZId: "env-zid"}, trx)
	require.NoError(t, err)
	f.ns, err = str.CreateNamespace(&store.Namespace{Token: "public", Name: "example.com", Open: true}, trx)
	require.NoError(t, err)
	require.NoError(t, trx.Commit())
	return f
}

type repairStoreSeed struct {
	mapping int
	name    int
	share   int
}

// seed creates a share with token holding name through a share name mapping, then marks the share or the
// name deleted as asked.
func (f *repairStoreFixture) seed(t *testing.T, token, name string, reserved, shareDeleted, nameDeleted bool) repairStoreSeed {
	t.Helper()
	trx, err := str.Begin()
	require.NoError(t, err)
	var s repairStoreSeed
	s.name, err = str.CreateName(&store.Name{NamespaceId: f.ns, Name: name, AccountId: f.acct, Reserved: reserved}, trx)
	require.NoError(t, err)
	s.share, err = str.CreateShare(f.env, &store.Share{ZId: name + "-zid", Token: token, ShareMode: "public", BackendMode: "proxy", PermissionMode: store.OpenPermissionMode}, trx)
	require.NoError(t, err)
	s.mapping, err = str.CreateShareNameMapping(&store.ShareNameMapping{ShareId: s.share, NameId: s.name}, trx)
	require.NoError(t, err)
	if shareDeleted {
		require.NoError(t, str.DeleteShare(s.share, trx))
	}
	if nameDeleted {
		require.NoError(t, str.DeleteName(s.name, trx))
	}
	require.NoError(t, trx.Commit())
	return s
}

func (f *repairStoreFixture) frontendMapping(t *testing.T, name, shareToken string) int {
	t.Helper()
	trx, err := str.Begin()
	require.NoError(t, err)
	id, err := str.CreateFrontendMapping(&store.FrontendMapping{FrontendToken: "dynamic-fe", Name: name, ShareToken: shareToken}, trx)
	require.NoError(t, err)
	require.NoError(t, trx.Commit())
	return id
}

func (f *repairStoreFixture) exec(t *testing.T, sql string) {
	t.Helper()
	trx, err := str.Begin()
	require.NoError(t, err)
	_, err = trx.Exec(sql)
	require.NoError(t, err)
	require.NoError(t, trx.Commit())
}

// repairStoreTables are the tables the sweep reads or writes.
var repairStoreTables = []string{"accounts", "environments", "shares", "names", "share_name_mappings", "frontend_mappings", "frontends", "access_grants"}

// repairStoreState dumps every row of the tables the sweep touches, for comparing before and after.
func repairStoreState(t *testing.T) string {
	t.Helper()
	return storeState(t, repairStoreTables...)
}

func mappingDeleted(t *testing.T, id int) bool {
	t.Helper()
	trx, err := str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()
	var deleted bool
	require.NoError(t, trx.QueryRow("select deleted from share_name_mappings where id = $1", id).Scan(&deleted))
	return deleted
}

func nameDeleted(t *testing.T, id int) bool {
	t.Helper()
	trx, err := str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()
	var deleted bool
	require.NoError(t, trx.QueryRow("select deleted from names where id = $1", id).Scan(&deleted))
	return deleted
}

func frontendMappingExists(t *testing.T, id int) bool {
	t.Helper()
	trx, err := str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()
	var count int
	require.NoError(t, trx.QueryRow("select count(*) from frontend_mappings where id = $1", id).Scan(&count))
	return count == 1
}

type repairStoreWorld struct {
	severedReserved, severedAuto, deadName, live, tornDown repairStoreSeed
	frontendDead, frontendGhost, frontendLive              int
}

// seedRepairStoreWorld seeds one row of each condition beside healthy rows: a live share holding a
// reserved name with a frontend mapping, and a share torn down properly (its mapping and auto-allocated
// name already soft-deleted).
func seedRepairStoreWorld(t *testing.T, f *repairStoreFixture) *repairStoreWorld {
	t.Helper()
	w := &repairStoreWorld{
		severedReserved: f.seed(t, "dead-res", "chosen", true, true, false),
		severedAuto:     f.seed(t, "dead-auto", "dead-auto", false, true, false),
		deadName:        f.seed(t, "live-dn", "gone", true, false, true),
		live:            f.seed(t, "live", "live", true, false, false),
		tornDown:        f.seed(t, "torn", "torn", false, true, true),
	}
	trx, err := str.Begin()
	require.NoError(t, err)
	require.NoError(t, str.DeleteShareNameMapping(w.tornDown.mapping, trx))
	require.NoError(t, trx.Commit())
	w.frontendDead = f.frontendMapping(t, "dead-auto.example.com", "dead-auto")
	w.frontendGhost = f.frontendMapping(t, "ghost.example.com", "ghost")
	w.frontendLive = f.frontendMapping(t, "live.example.com", "live")
	return w
}

func TestRepairStoreSurveyCountsAndSamples(t *testing.T) {
	f := setupRepairStoreFixture(t)
	w := seedRepairStoreWorld(t, f)

	rpt, err := repairStoreSurvey(RepairStoreOptions{Batch: DefaultRepairStoreBatch})
	require.NoError(t, err)

	found := make(map[string]int)
	samples := make(map[string][]repairStoreRow)
	for _, c := range rpt.conditions {
		found[c.name] = c.found
		for _, row := range c.sample {
			samples[c.name] = append(samples[c.name], *row)
		}
	}
	require.Equal(t, map[string]int{
		"environments of deleted accounts":       0,
		"shares in deleted environments":         0,
		"access frontends of deleted shares":     0,
		"names of deleted accounts":              0,
		"allocated names with no mapping":        0,
		"reserved names held by deleted shares":  1,
		"allocated names held by deleted shares": 1,
		"mappings to deleted names":              1,
		"frontend mappings without a live share": 2,
	}, found)
	require.Equal(t, 1, rpt.releasable)
	require.Equal(t, map[string][]repairStoreRow{
		"reserved names held by deleted shares":  {{id: int64(w.severedReserved.mapping), name: "chosen", scope: "public", shareToken: "dead-res"}},
		"allocated names held by deleted shares": {{id: int64(w.severedAuto.mapping), name: "dead-auto", scope: "public", shareToken: "dead-auto"}},
		"mappings to deleted names":              {{id: int64(w.deadName.mapping), name: "gone", scope: "public", shareToken: "live-dn"}},
		"frontend mappings without a live share": {
			{id: int64(w.frontendDead), name: "dead-auto.example.com", scope: "dynamic-fe", shareToken: "dead-auto"},
			{id: int64(w.frontendGhost), name: "ghost.example.com", scope: "dynamic-fe", shareToken: "ghost", noShareRow: true},
		},
	}, samples)

	out := &bytes.Buffer{}
	rpt.print(out)
	require.Contains(t, out.String(), "store repair dry run (nothing changed; run with --apply to repair)")
	require.Contains(t, out.String(), "allocated names to release: 1")
	require.Contains(t, out.String(), fmt.Sprintf("mapping '%d' name 'chosen' namespace 'public' share 'dead-res'", w.severedReserved.mapping))
	require.Contains(t, out.String(), fmt.Sprintf("mapping '%d' name 'ghost.example.com' frontend 'dynamic-fe' share 'ghost' (no share row)", w.frontendGhost))
	require.NotContains(t, out.String(), "'live'")
	require.NotContains(t, out.String(), "'torn'")
}

func TestRepairStoreSampleIsCapped(t *testing.T) {
	f := setupRepairStoreFixture(t)
	for i := 0; i < repairStoreSampleSize+5; i++ {
		f.seed(t, fmt.Sprintf("dead-%d", i), fmt.Sprintf("name-%d", i), true, true, false)
	}
	rpt, err := repairStoreSurvey(RepairStoreOptions{Batch: DefaultRepairStoreBatch})
	require.NoError(t, err)
	c := repairStoreConditionNamed(t, rpt, "reserved names held by deleted shares")
	require.Equal(t, repairStoreSampleSize+5, c.found)
	require.Len(t, c.sample, repairStoreSampleSize)
}

func TestRepairStoreDryRunChangesNothing(t *testing.T) {
	f := setupRepairStoreFixture(t)
	seedRepairStoreWorld(t, f)
	before := repairStoreState(t)

	out := &bytes.Buffer{}
	require.NoError(t, runRepairStore(RepairStoreOptions{Batch: DefaultRepairStoreBatch}, out))

	require.Equal(t, before, repairStoreState(t))
	require.NotContains(t, out.String(), "repaired")
}

func TestRepairStoreApplyRepairsMatchingRows(t *testing.T) {
	f := setupRepairStoreFixture(t)
	w := seedRepairStoreWorld(t, f)

	out := &bytes.Buffer{}
	require.NoError(t, runRepairStore(RepairStoreOptions{Apply: true, Batch: DefaultRepairStoreBatch}, out))

	// the severed mappings are released; the auto-allocated name goes with its mapping, the reserved one stays
	require.True(t, mappingDeleted(t, w.severedReserved.mapping))
	require.False(t, nameDeleted(t, w.severedReserved.name))
	require.True(t, mappingDeleted(t, w.severedAuto.mapping))
	require.True(t, nameDeleted(t, w.severedAuto.name))
	require.True(t, mappingDeleted(t, w.deadName.mapping))
	require.False(t, frontendMappingExists(t, w.frontendDead))
	require.False(t, frontendMappingExists(t, w.frontendGhost))

	// healthy rows are left alone
	require.False(t, mappingDeleted(t, w.live.mapping))
	require.False(t, nameDeleted(t, w.live.name))
	require.True(t, frontendMappingExists(t, w.frontendLive))

	require.Contains(t, out.String(), "names released: 1, batches: 4, failed batches: 0")
}

func TestRepairStoreBatches(t *testing.T) {
	f := setupRepairStoreFixture(t)
	for i := 0; i < 10; i++ {
		f.seed(t, fmt.Sprintf("dead-%d", i), fmt.Sprintf("name-%d", i), true, true, false)
	}

	out := &bytes.Buffer{}
	require.NoError(t, runRepairStore(RepairStoreOptions{Apply: true, Batch: 3}, out))

	require.Contains(t, out.String(), "names released: 0, batches: 4, failed batches: 0")
	logs := f.logs.String()
	require.Contains(t, logs, "batch '1' repaired '3' reserved names held by deleted shares ('0' names released); '3' done so far, '7' remaining")
	require.Contains(t, logs, "batch '2' repaired '3' reserved names held by deleted shares ('0' names released); '6' done so far, '4' remaining")
	require.Contains(t, logs, "batch '3' repaired '3' reserved names held by deleted shares ('0' names released); '9' done so far, '1' remaining")
	require.Contains(t, logs, "batch '4' repaired '1' reserved names held by deleted shares ('0' names released); '10' done so far, '0' remaining")

	rpt, err := repairStoreSurvey(RepairStoreOptions{Batch: 3})
	require.NoError(t, err)
	for _, c := range rpt.conditions {
		require.Zero(t, c.found, c.name)
	}
}

func TestRepairStoreOrderReservedFirst(t *testing.T) {
	f := setupRepairStoreFixture(t)
	// the auto-allocated mapping has the lower id, so only the condition order puts the reserved one first
	auto := f.seed(t, "dead-auto", "dead-auto", false, true, false)
	reserved := f.seed(t, "dead-res", "chosen", true, true, false)
	f.exec(t, "create table repair_order (seq integer primary key autoincrement, mapping_id integer not null)")
	f.exec(t, `create trigger record_repair_order after update of deleted on share_name_mappings when new.deleted and not old.deleted
		begin insert into repair_order (mapping_id) values (new.id); end`)

	require.NoError(t, runRepairStore(RepairStoreOptions{Apply: true, Batch: 1}, &bytes.Buffer{}))

	trx, err := str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()
	var order []int
	require.NoError(t, trx.Select(&order, "select mapping_id from repair_order order by seq"))
	require.Equal(t, []int{reserved.mapping, auto.mapping}, order)
}

func TestRepairStoreFailingBatchStops(t *testing.T) {
	f := setupRepairStoreFixture(t)
	var seeds []repairStoreSeed
	for i := 0; i < 7; i++ {
		seeds = append(seeds, f.seed(t, fmt.Sprintf("dead-%d", i), fmt.Sprintf("name-%d", i), true, true, false))
	}
	// the fifth mapping fails, in the second batch of three
	f.exec(t, fmt.Sprintf(`create trigger fail_repair before update on share_name_mappings when old.id = %d
		begin select raise(abort, 'injected repair failure'); end`, seeds[4].mapping))

	out := &bytes.Buffer{}
	err := runRepairStore(RepairStoreOptions{Apply: true, Batch: 3}, out)
	require.Error(t, err)
	require.Contains(t, err.Error(), "'1' batch failed")
	require.Contains(t, out.String(), "names released: 0, batches: 1, failed batches: 1")
	require.Contains(t, f.logs.String(), "batch '2' failed repairing reserved names held by deleted shares")

	// the first batch stays committed; the failing batch rolled back, and nothing after it ran
	for i, s := range seeds {
		require.Equal(t, i < 3, mappingDeleted(t, s.mapping), "mapping %d", i)
	}

	f.exec(t, "drop trigger fail_repair")
	out.Reset()
	require.NoError(t, runRepairStore(RepairStoreOptions{Apply: true, Batch: 3}, out))
	require.Contains(t, out.String(), "names released: 0, batches: 2, failed batches: 0")
	for i, s := range seeds {
		require.True(t, mappingDeleted(t, s.mapping), "mapping %d", i)
	}
}

func TestRepairStoreRejectsBatchBelowOne(t *testing.T) {
	setupRepairStoreFixture(t)
	require.Error(t, runRepairStore(RepairStoreOptions{Batch: 0}, &bytes.Buffer{}))
}

func repairStoreConditionNamed(t *testing.T, rpt *repairStoreReport, name string) *repairStoreCondition {
	t.Helper()
	for _, c := range rpt.conditions {
		if c.name == name {
			return c
		}
	}
	require.Failf(t, "no such condition", "%s", name)
	return nil
}

func (f *repairStoreFixture) inTrx(t *testing.T, fn func(trx *sqlx.Tx)) {
	t.Helper()
	trx, err := str.Begin()
	require.NoError(t, err)
	fn(trx)
	require.NoError(t, trx.Commit())
}

type strandedWorld struct {
	goneAccount, goneEnv, deadEnv                                int
	goneShare, stranded, oldShare, liveShare                     int
	goneShareName, goneReserved, leaked, liveReserved, liveAlloc int
	goneShareMapping, strandedMapping, liveMapping, liveAllocMap int
	staleAccess, liveAccess, strandedAccess                      int
	goneShareFrontend, strandedFrontend, liveFrontend            int
	strandedReserved                                             int
}

// seedStrandedWorld seeds conditions A to D beside a healthy live account with a live environment, a live
// share holding a reserved and an allocated name with a frontend mapping, and a live access to it:
//   - A: a deleted account with a live environment holding a live share on an allocated name with a frontend
//     mapping; and a deleted environment of the live account holding a live share on a reserved name with
//     a frontend mapping and an access
//   - B: an access, from the live environment, to a share deleted long ago
//   - C: a reserved name of the deleted account, held by nothing
//   - D: an allocated name of the live account held by nothing
func seedStrandedWorld(t *testing.T, f *repairStoreFixture) *strandedWorld {
	t.Helper()
	w := &strandedWorld{}
	f.inTrx(t, func(trx *sqlx.Tx) {
		var err error
		share := func(envID int, token string) int {
			id, err := str.CreateShare(envID, &store.Share{ZId: token + "-zid", Token: token, ShareMode: "public", BackendMode: "proxy", PermissionMode: store.OpenPermissionMode}, trx)
			require.NoError(t, err)
			return id
		}
		name := func(accountID int, n string, reserved bool) int {
			id, err := str.CreateName(&store.Name{NamespaceId: f.ns, Name: n, AccountId: accountID, Reserved: reserved}, trx)
			require.NoError(t, err)
			return id
		}
		mapping := func(shareID, nameID int) int {
			id, err := str.CreateShareNameMapping(&store.ShareNameMapping{ShareId: shareID, NameId: nameID}, trx)
			require.NoError(t, err)
			return id
		}
		frontendMapping := func(n, token string) int {
			id, err := str.CreateFrontendMapping(&store.FrontendMapping{FrontendToken: "dynamic-fe", Name: n + ".example.com", ShareToken: token}, trx)
			require.NoError(t, err)
			return id
		}
		access := func(envID int, token string, shareID int) int {
			id, err := str.CreateFrontend(envID, &store.Frontend{Token: token, ZId: token + "-zid", PrivateShareId: &shareID, PermissionMode: store.OpenPermissionMode}, trx)
			require.NoError(t, err)
			return id
		}

		// healthy
		w.liveShare = share(f.env, "live")
		w.liveReserved = name(f.acct, "live", true)
		w.liveMapping = mapping(w.liveShare, w.liveReserved)
		w.liveAlloc = name(f.acct, "live-alloc", false)
		w.liveAllocMap = mapping(w.liveShare, w.liveAlloc)
		w.liveFrontend = frontendMapping("live", "live")
		w.liveAccess = access(f.env, "live-access", w.liveShare)

		// A: a deleted account's live environment and its share
		w.goneAccount, err = str.CreateAccount(&store.Account{Email: "gone@example.com", Salt: "salt", Password: "password", Token: "gone-token"}, trx)
		require.NoError(t, err)
		w.goneEnv, err = str.CreateEnvironment(w.goneAccount, &store.Environment{Description: "gone", Host: "host", Address: "address", ZId: "gone-env-zid"}, trx)
		require.NoError(t, err)
		w.goneShare = share(w.goneEnv, "gone-share")
		w.goneShareName = name(w.goneAccount, "gone-share", false)
		w.goneShareMapping = mapping(w.goneShare, w.goneShareName)
		w.goneShareFrontend = frontendMapping("gone-share", "gone-share")
		require.NoError(t, str.DeleteAccount(w.goneAccount, trx))

		// A: a live account's deleted environment and its share
		w.deadEnv, err = str.CreateEnvironment(f.acct, &store.Environment{Description: "dead", Host: "host", Address: "address", ZId: "dead-env-zid"}, trx)
		require.NoError(t, err)
		w.stranded = share(w.deadEnv, "stranded")
		w.strandedReserved = name(f.acct, "stranded", true)
		w.strandedMapping = mapping(w.stranded, w.strandedReserved)
		w.strandedFrontend = frontendMapping("stranded", "stranded")
		w.strandedAccess = access(f.env, "stranded-access", w.stranded)
		require.NoError(t, str.DeleteEnvironment(w.deadEnv, trx))

		// B
		w.oldShare = share(f.env, "old-share")
		w.staleAccess = access(f.env, "stale-access", w.oldShare)
		require.NoError(t, str.DeleteShare(w.oldShare, trx))

		// C and D
		w.goneReserved = name(w.goneAccount, "gone-reserved", true)
		w.leaked = name(f.acct, "leaked", false)
	})
	return w
}

func frontendDeleted(t *testing.T, id int) bool {
	t.Helper()
	trx, err := str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()
	var deleted bool
	require.NoError(t, trx.QueryRow("select deleted from frontends where id = $1", id).Scan(&deleted))
	return deleted
}

func rowDeleted(t *testing.T, table string, id int) bool {
	t.Helper()
	trx, err := str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()
	var deleted bool
	require.NoError(t, trx.QueryRow("select deleted from "+table+" where id = $1", id).Scan(&deleted))
	return deleted
}

func TestRepairStoreStrandedSurvey(t *testing.T) {
	f := setupRepairStoreFixture(t)
	w := seedStrandedWorld(t, f)

	rpt, err := repairStoreSurvey(RepairStoreOptions{Batch: DefaultRepairStoreBatch})
	require.NoError(t, err)

	found := make(map[string]int)
	described := make(map[string][]string)
	for _, c := range rpt.conditions {
		found[c.name] = c.found
		for _, row := range c.sample {
			described[c.name] = append(described[c.name], row.describe(c.scope))
		}
	}
	require.Equal(t, map[string]int{
		"environments of deleted accounts":       1,
		"shares in deleted environments":         2,
		"access frontends of deleted shares":     1,
		"names of deleted accounts":              2,
		"allocated names with no mapping":        1,
		"reserved names held by deleted shares":  0,
		"allocated names held by deleted shares": 0,
		"mappings to deleted names":              0,
		"frontend mappings without a live share": 0,
	}, found)
	require.Equal(t, map[string][]string{
		"environments of deleted accounts": {fmt.Sprintf("environment '%d' zid 'gone-env-zid' account '%d'", w.goneEnv, w.goneAccount)},
		"shares in deleted environments": {
			fmt.Sprintf("share '%d' share 'gone-share' environment '%d'", w.goneShare, w.goneEnv),
			fmt.Sprintf("share '%d' share 'stranded' environment '%d'", w.stranded, w.deadEnv),
		},
		"access frontends of deleted shares": {fmt.Sprintf("frontend '%d' token 'stale-access' share 'old-share' environment '%d'", w.staleAccess, f.env)},
		"names of deleted accounts": {
			fmt.Sprintf("name '%d' name 'gone-share' namespace 'public' account '%d'", w.goneShareName, w.goneAccount),
			fmt.Sprintf("name '%d' name 'gone-reserved' namespace 'public' account '%d'", w.goneReserved, w.goneAccount),
		},
		"allocated names with no mapping": {fmt.Sprintf("name '%d' name 'leaked' namespace 'public' account '%d'", w.leaked, f.acct)},
	}, described)
}

func TestRepairStoreStrandedDryRunChangesNothing(t *testing.T) {
	f := setupRepairStoreFixture(t)
	seedStrandedWorld(t, f)
	before := repairStoreState(t)

	out := &bytes.Buffer{}
	require.NoError(t, runRepairStore(RepairStoreOptions{Batch: DefaultRepairStoreBatch}, out))

	require.Equal(t, before, repairStoreState(t))
	require.NotContains(t, out.String(), "left in OpenZiti")
}

func TestRepairStoreStrandedApply(t *testing.T) {
	f := setupRepairStoreFixture(t)
	w := seedStrandedWorld(t, f)

	out := &bytes.Buffer{}
	require.NoError(t, runRepairStore(RepairStoreOptions{Apply: true, Batch: DefaultRepairStoreBatch}, out))

	// A: the environment and both shares, with what their teardown would have released
	require.True(t, rowDeleted(t, "environments", w.goneEnv))
	require.True(t, rowDeleted(t, "shares", w.goneShare))
	require.True(t, mappingDeleted(t, w.goneShareMapping))
	require.True(t, nameDeleted(t, w.goneShareName))
	require.False(t, frontendMappingExists(t, w.goneShareFrontend))
	require.True(t, rowDeleted(t, "shares", w.stranded))
	require.True(t, mappingDeleted(t, w.strandedMapping))
	require.False(t, nameDeleted(t, w.strandedReserved), "a live account's reserved name survives its share")
	require.False(t, frontendMappingExists(t, w.strandedFrontend))
	require.True(t, frontendDeleted(t, w.strandedAccess))
	// B, C, D
	require.True(t, frontendDeleted(t, w.staleAccess))
	require.True(t, nameDeleted(t, w.goneReserved))
	require.True(t, nameDeleted(t, w.leaked))

	// healthy rows survive
	require.False(t, rowDeleted(t, "accounts", f.acct))
	require.False(t, rowDeleted(t, "environments", f.env))
	require.False(t, rowDeleted(t, "shares", w.liveShare))
	require.False(t, mappingDeleted(t, w.liveMapping))
	require.False(t, mappingDeleted(t, w.liveAllocMap))
	require.False(t, nameDeleted(t, w.liveReserved))
	require.False(t, nameDeleted(t, w.liveAlloc))
	require.True(t, frontendMappingExists(t, w.liveFrontend))
	require.False(t, frontendDeleted(t, w.liveAccess))

	// in order: one batch per condition that found anything, A to D; the deleted account's allocated
	// name was released with its share under A, so C repairs only the reserved one
	logs := f.logs.String()
	var at []int
	for _, line := range []string{
		"batch '1' repaired '1' environments of deleted accounts",
		"batch '2' repaired '2' shares in deleted environments",
		"batch '3' repaired '1' access frontends of deleted shares",
		"batch '4' repaired '1' names of deleted accounts",
		"batch '5' repaired '1' allocated names with no mapping",
	} {
		i := strings.Index(logs, line)
		require.GreaterOrEqual(t, i, 0, line)
		at = append(at, i)
	}
	require.IsIncreasing(t, at)
	require.Contains(t, out.String(), "batches: 5, failed batches: 0")
	require.Contains(t, out.String(), "\nidentities left in OpenZiti by environments of deleted accounts (1); remove each with 'zrok2 admin delete identity <zId>':\n  'gone-env-zid'\n")

	rpt, err := repairStoreSurvey(RepairStoreOptions{Batch: DefaultRepairStoreBatch})
	require.NoError(t, err)
	for _, c := range rpt.conditions {
		require.Zero(t, c.found, c.name)
	}
}

// a share in a deleted environment holding a reserved name with a frontend mapping is repaired whole by
// condition A in one run, and nothing is left for the mapping conditions after it.
func TestRepairStoreStrandedShareRepairedWhole(t *testing.T) {
	f := setupRepairStoreFixture(t)
	var envID, shareID int
	s := repairStoreSeed{}
	f.inTrx(t, func(trx *sqlx.Tx) {
		var err error
		envID, err = str.CreateEnvironment(f.acct, &store.Environment{Description: "dead", Host: "host", Address: "address", ZId: "dead-env-zid"}, trx)
		require.NoError(t, err)
		shareID, err = str.CreateShare(envID, &store.Share{ZId: "stranded-zid", Token: "stranded", ShareMode: "public", BackendMode: "proxy", PermissionMode: store.OpenPermissionMode}, trx)
		require.NoError(t, err)
		s.name, err = str.CreateName(&store.Name{NamespaceId: f.ns, Name: "chosen", AccountId: f.acct, Reserved: true}, trx)
		require.NoError(t, err)
		s.mapping, err = str.CreateShareNameMapping(&store.ShareNameMapping{ShareId: shareID, NameId: s.name}, trx)
		require.NoError(t, err)
		require.NoError(t, str.DeleteEnvironment(envID, trx))
	})
	fm := f.frontendMapping(t, "chosen.example.com", "stranded")

	out := &bytes.Buffer{}
	require.NoError(t, runRepairStore(RepairStoreOptions{Apply: true, Batch: 1}, out))

	require.True(t, rowDeleted(t, "shares", shareID))
	require.True(t, mappingDeleted(t, s.mapping))
	require.False(t, frontendMappingExists(t, fm))
	require.False(t, nameDeleted(t, s.name))
	require.Contains(t, out.String(), "batches: 1, failed batches: 0")

	rpt, err := repairStoreSurvey(RepairStoreOptions{Batch: 1})
	require.NoError(t, err)
	for _, c := range rpt.conditions {
		require.Zero(t, c.found, c.name)
	}
}

// an environment whose batch rolled back is still in the store, so it is not listed as left in ziti.
func TestRepairStoreListsOnlyCommittedIdentities(t *testing.T) {
	f := setupRepairStoreFixture(t)
	var envIDs []int
	f.inTrx(t, func(trx *sqlx.Tx) {
		goneID, err := str.CreateAccount(&store.Account{Email: "gone@example.com", Salt: "salt", Password: "password", Token: "gone-token"}, trx)
		require.NoError(t, err)
		for _, zId := range []string{"gone-env-1", "gone-env-2"} {
			id, err := str.CreateEnvironment(goneID, &store.Environment{Description: zId, Host: "host", Address: "address", ZId: zId}, trx)
			require.NoError(t, err)
			envIDs = append(envIDs, id)
		}
		require.NoError(t, str.DeleteAccount(goneID, trx))
	})
	f.exec(t, fmt.Sprintf(`create trigger fail_env before update on environments when old.id = %d
		begin select raise(abort, 'injected repair failure'); end`, envIDs[1]))

	out := &bytes.Buffer{}
	require.Error(t, runRepairStore(RepairStoreOptions{Apply: true, Batch: 1}, out))

	require.Contains(t, out.String(), "identities left in OpenZiti by environments of deleted accounts (1); remove each with 'zrok2 admin delete identity <zId>':\n  'gone-env-1'\n")
	require.NotContains(t, out.String(), "'gone-env-2'\n")
	require.True(t, rowDeleted(t, "environments", envIDs[0]))
	require.False(t, rowDeleted(t, "environments", envIDs[1]))
}
