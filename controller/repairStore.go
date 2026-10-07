package controller

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/jmoiron/sqlx"
	"github.com/michaelquigley/df/dl"
	zrok_config "github.com/openziti/zrok/v2/controller/config"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/pkg/errors"
)

// DefaultRepairStoreBatch is how many rows one repair transaction touches, small enough that no batch
// holds a long transaction on the store.
const DefaultRepairStoreBatch = 500

// repairStoreSampleSize is how many rows of each condition the survey lists for spot-checking.
const repairStoreSampleSize = 20

type RepairStoreOptions struct {
	// Apply repairs the rows; without it the sweep only reports.
	Apply bool
	// Batch is how many rows each repair transaction touches.
	Batch int
}

// repairStoreRow is one matching row, as the survey sample lists it. kind names the row's table ("mapping"
// when empty); scope is the value the condition's scope label names, such as a namespace token, a frontend
// token or an environment's ziti id; owner names the environment or account of a row with no name.
type repairStoreRow struct {
	kind       string
	id         int64
	name       string
	scope      string
	shareToken string
	noShareRow bool
	owner      string
}

// repairStoreCondition is one kind of row a teardown would have released had it run.
// repairStoreConditions returns them in repair order.
type repairStoreCondition struct {
	name   string
	scope  string
	count  func(trx *sqlx.Tx) (int, error)
	find   func(limit int, trx *sqlx.Tx) ([]*repairStoreRow, error)
	// repair runs one batch. left names what the batch released from the store but left in ziti: the
	// identities of the environments it soft-deleted.
	repair func(limit int, trx *sqlx.Tx) (repaired, released int, left []string, err error)

	found    int
	sample   []*repairStoreRow
	repaired int
	left     []string
}

// repairStoreConditions lists the conditions in repair order. the stranded environments and shares come
// first, since releasing a share's store rows repairs much of what the later conditions would find; then
// access frontends of deleted shares, names of deleted accounts and unmapped allocated names. then the
// share name mappings: reserved names first, so the names users chose are freed first if a run is
// interrupted, then auto-allocated names with the names themselves, then mappings to deleted names, then
// frontend mappings.
func repairStoreConditions() []*repairStoreCondition {
	return []*repairStoreCondition{
		{
			name:  "environments of deleted accounts",
			scope: "zid",
			count: str.CountEnvironmentsOfDeletedAccounts,
			find: func(limit int, trx *sqlx.Tx) ([]*repairStoreRow, error) {
				envs, err := str.FindEnvironmentsOfDeletedAccounts(limit, trx)
				if err != nil {
					return nil, err
				}
				rows := make([]*repairStoreRow, 0, len(envs))
				for _, env := range envs {
					rows = append(rows, &repairStoreRow{kind: "environment", id: int64(env.Id), scope: env.ZId, owner: repairStoreOwner("account", env.AccountId)})
				}
				return rows, nil
			},
			repair: func(limit int, trx *sqlx.Tx) (int, int, []string, error) {
				envs, err := str.FindEnvironmentsOfDeletedAccounts(limit, trx)
				if err != nil {
					return 0, 0, nil, err
				}
				var zIds []string
				for _, env := range envs {
					if err := removeEnvironmentFromStore(env, trx); err != nil {
						return 0, 0, nil, err
					}
					zIds = append(zIds, env.ZId)
				}
				return len(envs), 0, zIds, nil
			},
		},
		{
			name:  "shares in deleted environments",
			count: str.CountStrandedShares,
			find: func(limit int, trx *sqlx.Tx) ([]*repairStoreRow, error) {
				shrs, err := str.FindStrandedShares(limit, trx)
				if err != nil {
					return nil, err
				}
				rows := make([]*repairStoreRow, 0, len(shrs))
				for _, shr := range shrs {
					rows = append(rows, &repairStoreRow{kind: "share", id: int64(shr.Id), shareToken: shr.Token, owner: repairStoreOwner("environment", &shr.EnvironmentId)})
				}
				return rows, nil
			},
			repair: func(limit int, trx *sqlx.Tx) (int, int, []string, error) {
				shrs, err := str.FindStrandedShares(limit, trx)
				if err != nil {
					return 0, 0, nil, err
				}
				// the store half of the teardown only; the unbinds are not published, and the share's ziti
				// objects are left for gc, which collects them once the token is no longer live
				for _, shr := range shrs {
					if _, err := releaseShareFromStore(shr, trx); err != nil {
						return 0, 0, nil, err
					}
				}
				return len(shrs), 0, nil, nil
			},
		},
		{
			name:  "access frontends of deleted shares",
			scope: "token",
			count: str.CountAccessFrontendsOfDeletedShares,
			find: func(limit int, trx *sqlx.Tx) ([]*repairStoreRow, error) {
				fes, err := str.FindAccessFrontendsOfDeletedShares(limit, trx)
				if err != nil {
					return nil, err
				}
				rows := make([]*repairStoreRow, 0, len(fes))
				for _, fe := range fes {
					rows = append(rows, &repairStoreRow{kind: "frontend", id: int64(fe.Id), scope: fe.Token, shareToken: fe.ShareToken, owner: repairStoreOwner("environment", fe.EnvironmentId)})
				}
				return rows, nil
			},
			repair: func(limit int, trx *sqlx.Tx) (int, int, []string, error) {
				fes, err := str.FindAccessFrontendsOfDeletedShares(limit, trx)
				if err != nil {
					return 0, 0, nil, err
				}
				ids := make([]int, 0, len(fes))
				for _, fe := range fes {
					ids = append(ids, fe.Id)
				}
				n, err := str.DeleteFrontends(ids, trx)
				return int(n), 0, nil, err
			},
		},
		{
			name:  "names of deleted accounts",
			scope: "namespace",
			count: str.CountNamesOfDeletedAccounts,
			find: func(limit int, trx *sqlx.Tx) ([]*repairStoreRow, error) {
				return repairStoreNameRows(str.FindNamesOfDeletedAccounts(limit, trx))
			},
			repair: func(limit int, trx *sqlx.Tx) (int, int, []string, error) {
				ids, err := repairStoreNameIds(str.FindNamesOfDeletedAccounts(limit, trx))
				if err != nil {
					return 0, 0, nil, err
				}
				if _, err := str.DeleteShareNameMappingsForNames(ids, trx); err != nil {
					return 0, 0, nil, err
				}
				n, err := str.DeleteNames(ids, trx)
				return int(n), 0, nil, err
			},
		},
		{
			name:  "allocated names with no mapping",
			scope: "namespace",
			count: str.CountUnmappedAllocatedNames,
			find: func(limit int, trx *sqlx.Tx) ([]*repairStoreRow, error) {
				return repairStoreNameRows(str.FindUnmappedAllocatedNames(limit, trx))
			},
			repair: func(limit int, trx *sqlx.Tx) (int, int, []string, error) {
				ids, err := repairStoreNameIds(str.FindUnmappedAllocatedNames(limit, trx))
				if err != nil {
					return 0, 0, nil, err
				}
				n, err := str.DeleteUnmappedAllocatedNames(ids, trx)
				return int(n), 0, nil, err
			},
		},
		{
			name:  "reserved names held by deleted shares",
			scope: "namespace",
			count: func(trx *sqlx.Tx) (int, error) { return str.CountShareNameMappingsToDeletedShares(true, trx) },
			find: func(limit int, trx *sqlx.Tx) ([]*repairStoreRow, error) {
				return repairStoreMappingRows(str.FindShareNameMappingsToDeletedShares(true, limit, trx))
			},
			repair: func(limit int, trx *sqlx.Tx) (int, int, []string, error) {
				return noZitiLeftovers(repairStoreMappings(trx, false)(str.FindShareNameMappingsToDeletedShares(true, limit, trx)))
			},
		},
		{
			name:  "allocated names held by deleted shares",
			scope: "namespace",
			count: func(trx *sqlx.Tx) (int, error) { return str.CountShareNameMappingsToDeletedShares(false, trx) },
			find: func(limit int, trx *sqlx.Tx) ([]*repairStoreRow, error) {
				return repairStoreMappingRows(str.FindShareNameMappingsToDeletedShares(false, limit, trx))
			},
			repair: func(limit int, trx *sqlx.Tx) (int, int, []string, error) {
				return noZitiLeftovers(repairStoreMappings(trx, true)(str.FindShareNameMappingsToDeletedShares(false, limit, trx)))
			},
		},
		{
			name:  "mappings to deleted names",
			scope: "namespace",
			count: str.CountShareNameMappingsToDeletedNames,
			find: func(limit int, trx *sqlx.Tx) ([]*repairStoreRow, error) {
				return repairStoreMappingRows(str.FindShareNameMappingsToDeletedNames(limit, trx))
			},
			repair: func(limit int, trx *sqlx.Tx) (int, int, []string, error) {
				return noZitiLeftovers(repairStoreMappings(trx, false)(str.FindShareNameMappingsToDeletedNames(limit, trx)))
			},
		},
		{
			name:  "frontend mappings without a live share",
			scope: "frontend",
			count: str.CountFrontendMappingsWithoutLiveShare,
			find: func(limit int, trx *sqlx.Tx) ([]*repairStoreRow, error) {
				fms, err := str.FindFrontendMappingsWithoutLiveShare(limit, trx)
				if err != nil {
					return nil, err
				}
				rows := make([]*repairStoreRow, 0, len(fms))
				for _, fm := range fms {
					rows = append(rows, &repairStoreRow{id: fm.Id, name: fm.Name, scope: fm.FrontendToken, shareToken: fm.ShareToken, noShareRow: !fm.ShareRow})
				}
				return rows, nil
			},
			repair: func(limit int, trx *sqlx.Tx) (int, int, []string, error) {
				fms, err := str.FindFrontendMappingsWithoutLiveShare(limit, trx)
				if err != nil {
					return 0, 0, nil, err
				}
				ids := make([]int64, 0, len(fms))
				for _, fm := range fms {
					ids = append(ids, fm.Id)
				}
				n, err := str.DeleteFrontendMappings(ids, trx)
				return int(n), 0, nil, err
			},
		},
	}
}

// noZitiLeftovers adapts a repair that leaves nothing in ziti to the repair signature.
func noZitiLeftovers(repaired, released int, err error) (int, int, []string, error) {
	return repaired, released, nil, err
}

func repairStoreOwner(kind string, id *int) string {
	if id == nil {
		return ""
	}
	return fmt.Sprintf("%s '%d'", kind, *id)
}

func repairStoreNameRows(names []*store.NameRepairDetail, err error) ([]*repairStoreRow, error) {
	if err != nil {
		return nil, err
	}
	rows := make([]*repairStoreRow, 0, len(names))
	for _, n := range names {
		rows = append(rows, &repairStoreRow{kind: "name", id: int64(n.Id), name: n.Name, scope: n.NamespaceToken, owner: repairStoreOwner("account", &n.AccountId)})
	}
	return rows, nil
}

func repairStoreNameIds(names []*store.NameRepairDetail, err error) ([]int, error) {
	if err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(names))
	for _, n := range names {
		ids = append(ids, n.Id)
	}
	return ids, nil
}

func repairStoreMappingRows(details []*store.ShareNameMappingRepairDetail, err error) ([]*repairStoreRow, error) {
	if err != nil {
		return nil, err
	}
	rows := make([]*repairStoreRow, 0, len(details))
	for _, d := range details {
		rows = append(rows, &repairStoreRow{id: int64(d.MappingId), name: d.Name, scope: d.NamespaceToken, shareToken: d.ShareToken})
	}
	return rows, nil
}

// repairStoreMappings returns a repair that soft-deletes the selected share name mappings by id and, with
// release, the live auto-allocated names they held, as teardownShare does.
func repairStoreMappings(trx *sqlx.Tx, release bool) func([]*store.ShareNameMappingRepairDetail, error) (int, int, error) {
	return func(details []*store.ShareNameMappingRepairDetail, err error) (int, int, error) {
		if err != nil {
			return 0, 0, err
		}
		mappingIds := make([]int, 0, len(details))
		var nameIds []int
		for _, d := range details {
			mappingIds = append(mappingIds, d.MappingId)
			if release && !d.Reserved && !d.NameDeleted {
				nameIds = append(nameIds, d.NameId)
			}
		}
		repaired, err := str.DeleteShareNameMappings(mappingIds, trx)
		if err != nil {
			return 0, 0, err
		}
		released, err := str.DeleteAllocatedNames(nameIds, trx)
		if err != nil {
			return 0, 0, err
		}
		return int(repaired), int(released), nil
	}
}

type repairStoreReport struct {
	apply         bool
	batch         int
	conditions    []*repairStoreCondition
	releasable    int
	released      int
	batches       int
	failedBatches int
}

// RepairStore reports, and with opts.Apply repairs, the store rows a teardown would have released had it
// run: stranded environments and shares, access frontends of deleted shares, names of deleted accounts,
// unmapped allocated names, and share name and frontend mappings of deleted shares and names. it touches
// the store only: no ziti calls and no frontend notifications.
func RepairStore(inCfg *zrok_config.Config, opts RepairStoreOptions) error {
	cfg = inCfg
	if v, err := openAdminStore(cfg.Store); err == nil {
		str = v
	} else {
		return err
	}
	defer func() {
		if err := str.Close(); err != nil {
			dl.Errorf("error closing store: %v", err)
		}
	}()
	return runRepairStore(opts, os.Stdout)
}

func runRepairStore(opts RepairStoreOptions, out io.Writer) error {
	if opts.Batch < 1 {
		return errors.Errorf("batch size must be at least 1, not '%d'", opts.Batch)
	}
	rpt, err := repairStoreSurvey(opts)
	if err != nil {
		return err
	}
	rpt.print(out)
	if !opts.Apply {
		return nil
	}
	err = repairStoreApply(rpt)
	rpt.printRepaired(out)
	return err
}

// repairStoreSurvey counts and samples every condition in one transaction, which it rolls back.
func repairStoreSurvey(opts RepairStoreOptions) (*repairStoreReport, error) {
	trx, err := str.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "error starting transaction")
	}
	defer func() { _ = trx.Rollback() }()

	rpt := &repairStoreReport{apply: opts.Apply, batch: opts.Batch, conditions: repairStoreConditions()}
	for _, c := range rpt.conditions {
		if c.found, err = c.count(trx); err != nil {
			return nil, errors.Wrapf(err, "error counting %s", c.name)
		}
		if c.sample, err = c.find(repairStoreSampleSize, trx); err != nil {
			return nil, errors.Wrapf(err, "error sampling %s", c.name)
		}
		dl.Infof("found '%d' %s", c.found, c.name)
	}
	if rpt.releasable, err = str.CountAllocatedNamesHeldByDeletedShares(trx); err != nil {
		return nil, errors.Wrap(err, "error counting releasable names")
	}
	dl.Infof("found '%d' allocated names to release", rpt.releasable)
	return rpt, nil
}

// repairStoreApply repairs each condition in order, a batch per transaction, until a batch finds nothing.
// a failed batch rolls back and stops the run; the batches before it stay committed, and a later run picks
// up where it stopped, since a repaired row no longer matches.
func repairStoreApply(rpt *repairStoreReport) error {
	for _, c := range rpt.conditions {
		for {
			repaired, released, left, err := repairStoreBatch(c, rpt.batch)
			if err != nil {
				rpt.failedBatches++
				dl.Errorf("batch '%d' failed repairing %s; its changes were rolled back and the run stops: %v", rpt.batches+1, c.name, err)
				return errors.Wrapf(err, "'%d' batch failed repairing %s; see the log", rpt.failedBatches, c.name)
			}
			if repaired == 0 {
				break
			}
			rpt.batches++
			c.repaired += repaired
			c.left = append(c.left, left...)
			rpt.released += released
			dl.Infof("batch '%d' repaired '%d' %s ('%d' names released); '%d' done so far, '%d' remaining", rpt.batches, repaired, c.name, released, c.repaired, max(c.found-c.repaired, 0))
		}
	}
	dl.Infof("repaired %s in '%d' batches; released '%d' names", rpt.repairedSummary(), rpt.batches, rpt.released)
	return nil
}

// repairStoreBatch runs one batch of c in its own transaction. what it reports, the ziti leftovers
// included, counts only once the batch has committed.
func repairStoreBatch(c *repairStoreCondition, batch int) (int, int, []string, error) {
	trx, err := str.Begin()
	if err != nil {
		return 0, 0, nil, errors.Wrap(err, "error starting transaction")
	}
	defer func() { _ = trx.Rollback() }()
	repaired, released, left, err := c.repair(batch, trx)
	if err != nil {
		return 0, 0, nil, err
	}
	if repaired == 0 {
		return 0, 0, nil, nil
	}
	if err := trx.Commit(); err != nil {
		return 0, 0, nil, errors.Wrap(err, "error committing batch")
	}
	return repaired, released, left, nil
}

func (rpt *repairStoreReport) repairedSummary() string {
	s := ""
	for i, c := range rpt.conditions {
		if i > 0 {
			s += ", "
		}
		s += fmt.Sprintf("'%d' %s", c.repaired, c.name)
	}
	return s
}

func (rpt *repairStoreReport) print(out io.Writer) {
	if rpt.apply {
		_, _ = fmt.Fprintf(out, "store repair (repairing in batches of %d)\n\n", rpt.batch)
	} else {
		_, _ = fmt.Fprintf(out, "store repair dry run (nothing changed; run with --apply to repair)\n\n")
	}

	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintf(tw, "condition\tfound\t\n")
	for _, c := range rpt.conditions {
		_, _ = fmt.Fprintf(tw, "%s\t%d\t\n", c.name, c.found)
	}
	_ = tw.Flush()
	_, _ = fmt.Fprintf(out, "allocated names to release: %d\n", rpt.releasable)

	sampled := false
	for _, c := range rpt.conditions {
		if len(c.sample) == 0 {
			continue
		}
		sampled = true
		_, _ = fmt.Fprintf(out, "\n%s (%d of %d):\n", c.name, len(c.sample), c.found)
		for _, row := range c.sample {
			_, _ = fmt.Fprintf(out, "  %s\n", row.describe(c.scope))
		}
	}
	if !sampled {
		_, _ = fmt.Fprintf(out, "\nnothing to repair\n")
	}
}

func (rpt *repairStoreReport) printRepaired(out io.Writer) {
	_, _ = fmt.Fprintf(out, "\n")
	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintf(tw, "condition\tfound\trepaired\t\n")
	for _, c := range rpt.conditions {
		_, _ = fmt.Fprintf(tw, "%s\t%d\t%d\t\n", c.name, c.found, c.repaired)
	}
	_ = tw.Flush()
	_, _ = fmt.Fprintf(out, "names released: %d, batches: %d, failed batches: %d\n", rpt.released, rpt.batches, rpt.failedBatches)

	for _, c := range rpt.conditions {
		if len(c.left) == 0 {
			continue
		}
		_, _ = fmt.Fprintf(out, "\nidentities left in OpenZiti by %s (%d); remove each with 'zrok2 admin delete identity <zId>':\n", c.name, len(c.left))
		for _, zId := range c.left {
			_, _ = fmt.Fprintf(out, "  '%s'\n", zId)
		}
	}
}

// describe formats a sample row: its kind and id, then whichever of name, scope, share and owner it has.
func (row *repairStoreRow) describe(scopeLabel string) string {
	kind := row.kind
	if kind == "" {
		kind = "mapping"
	}
	line := fmt.Sprintf("%s '%d'", kind, row.id)
	if row.name != "" {
		line += fmt.Sprintf(" name '%s'", row.name)
	}
	if row.scope != "" {
		line += fmt.Sprintf(" %s '%s'", scopeLabel, row.scope)
	}
	if row.shareToken != "" {
		line += fmt.Sprintf(" share '%s'", row.shareToken)
	}
	if row.noShareRow {
		line += " (no share row)"
	}
	if row.owner != "" {
		line += " " + row.owner
	}
	return line
}
