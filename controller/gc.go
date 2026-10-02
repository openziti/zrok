package controller

import (
	"fmt"
	"io"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/zrok/v2/controller/automation"
	zrok_config "github.com/openziti/zrok/v2/controller/config"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/pkg/errors"
)

// DefaultGCMinAge is how old an orphaned object must be before gc touches it. a day is well clear of
// clock skew between the ziti and zrok hosts, and with the leak stopped nothing younger needs reclaiming.
// any guard at all also keeps gc off a share whose
// ziti objects are allocated but whose row has not yet committed is never collected.
const DefaultGCMinAge = 24 * time.Hour

type GCOptions struct {
	// Delete removes the orphaned objects; without it gc only reports.
	Delete bool
	// MinAge skips orphaned objects younger than this by their createdAt.
	MinAge time.Duration
	// pageSize overrides automation.MaxPageSize for tests.
	pageSize int64
}

// gcFilter lists everything zrok tagged; ownership is then decided by the zrokShareToken tag.
const gcFilter = "tags.zrok != null"

const gcShareTokenTag = "zrokShareToken"

// gcKind is one kind of ziti object gc collects. gcKinds returns them in deletion order.
type gcKind struct {
	name   string
	list   func(pageSize int64) ([]*gcObject, error)
	delete func(id string) error
}

// gcKinds lists the kinds in the order their orphans are deleted: policies and configs before services,
// so a service is never left referenced by an object that outlives it.
func gcKinds(ziti *automation.ZitiAutomation) []*gcKind {
	return []*gcKind{
		{
			name: "service edge router policies",
			list: func(pageSize int64) ([]*gcObject, error) {
				return gcList(ziti.ServiceEdgeRouterPolicies.Find, pageSize, func(v *rest_model.ServiceEdgeRouterPolicyDetail) (*rest_model.BaseEntity, *string) {
					return &v.BaseEntity, v.Name
				})
			},
			delete: ziti.ServiceEdgeRouterPolicies.Delete,
		},
		{
			name: "service policies",
			list: func(pageSize int64) ([]*gcObject, error) {
				return gcList(ziti.ServicePolicies.Find, pageSize, func(v *rest_model.ServicePolicyDetail) (*rest_model.BaseEntity, *string) {
					return &v.BaseEntity, v.Name
				})
			},
			delete: ziti.ServicePolicies.Delete,
		},
		{
			name: "configs",
			list: func(pageSize int64) ([]*gcObject, error) {
				return gcList(ziti.Configs.Find, pageSize, func(v *rest_model.ConfigDetail) (*rest_model.BaseEntity, *string) {
					return &v.BaseEntity, v.Name
				})
			},
			delete: ziti.Configs.Delete,
		},
		{
			name: "services",
			list: func(pageSize int64) ([]*gcObject, error) {
				return gcList(ziti.Services.Find, pageSize, func(v *rest_model.ServiceDetail) (*rest_model.BaseEntity, *string) {
					return &v.BaseEntity, v.Name
				})
			},
			delete: ziti.Services.Delete,
		},
	}
}

type gcObject struct {
	id         string
	name       string
	shareToken string
	createdAt  time.Time
	hasCreated bool
}

func gcList[T any](finder func(*automation.FilterOptions) ([]*T, error), pageSize int64, entity func(*T) (*rest_model.BaseEntity, *string)) ([]*gcObject, error) {
	items, err := automation.FindAll(finder, gcFilter, pageSize)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool, len(items))
	out := make([]*gcObject, 0, len(items))
	for _, item := range items {
		base, name := entity(item)
		if base.ID == nil || seen[*base.ID] {
			// an offset listing can repeat an object when another process deletes concurrently.
			continue
		}
		seen[*base.ID] = true
		obj := &gcObject{id: *base.ID}
		if name != nil {
			obj.name = *name
		}
		if base.Tags != nil {
			if token, ok := base.Tags.SubTags[gcShareTokenTag].(string); ok {
				obj.shareToken = token
			}
		}
		if base.CreatedAt != nil {
			obj.createdAt = time.Time(*base.CreatedAt)
			obj.hasCreated = true
		}
		out = append(out, obj)
	}
	return out, nil
}

type gcKindReport struct {
	kind     *gcKind
	live     int
	tooYoung int
	unowned  int
	orphans  []*gcObject

	deleted     int
	alreadyGone int
	failed      int
}

func (r *gcKindReport) total() int {
	return r.live + len(r.orphans) + r.tooYoung + r.unowned
}

type gcReport struct {
	liveShares int
	minAge     time.Duration
	deleteMode bool
	kinds      []*gcKindReport
}

// GC reports, and with opts.Delete removes, the ziti objects carrying a zrokShareToken tag whose token
// belongs to no live share.
func GC(inCfg *zrok_config.Config, opts GCOptions) error {
	cfg = inCfg
	if v, err := store.Open(cfg.Store); err == nil {
		str = v
	} else {
		return errors.Wrap(err, "error opening store")
	}
	defer func() {
		if err := str.Close(); err != nil {
			dl.Errorf("error closing store: %v", err)
		}
	}()
	ziti, err := automation.NewZitiAutomation(cfg.Ziti)
	if err != nil {
		return err
	}
	return runGC(ziti, opts, os.Stdout)
}

func runGC(ziti *automation.ZitiAutomation, opts GCOptions, out io.Writer) error {
	rpt, err := gcSurvey(ziti, opts)
	if err != nil {
		return err
	}
	rpt.print(out)
	if !opts.Delete {
		return nil
	}
	failed := gcReclaim(rpt)
	rpt.printDeleted(out)
	if failed > 0 {
		return errors.Errorf("'%d' deletes failed; see the log", failed)
	}
	return nil
}

// gcSurvey reads the live share tokens and then every zrok-tagged object of each kind, and sorts each
// object into live, orphaned, too young or not owned by a share. the store is read first, so an object
// created after the read is younger than the run and falls under the age guard.
func gcSurvey(ziti *automation.ZitiAutomation, opts GCOptions) (*gcReport, error) {
	live, err := gcLiveTokens()
	if err != nil {
		return nil, err
	}
	rpt := &gcReport{liveShares: len(live), minAge: opts.MinAge, deleteMode: opts.Delete}
	now := time.Now()
	for _, kind := range gcKinds(ziti) {
		objs, err := kind.list(opts.pageSize)
		if err != nil {
			return nil, errors.Wrapf(err, "error listing %s", kind.name)
		}
		kr := &gcKindReport{kind: kind}
		for _, obj := range objs {
			switch {
			case obj.shareToken == "":
				kr.unowned++
			case live[obj.shareToken]:
				kr.live++
				dl.Debugf("live %s '%s' ('%s'), share '%s'", kind.name, obj.id, obj.name, obj.shareToken)
			case !obj.hasCreated || now.Sub(obj.createdAt) < opts.MinAge:
				kr.tooYoung++
				dl.Infof("too young to collect: %s '%s' ('%s'), share '%s', created '%s'", kind.name, obj.id, obj.name, obj.shareToken, obj.createdAt.Format(time.RFC3339))
			default:
				kr.orphans = append(kr.orphans, obj)
				dl.Infof("orphaned: %s '%s' ('%s'), share '%s', created '%s'", kind.name, obj.id, obj.name, obj.shareToken, obj.createdAt.Format(time.RFC3339))
			}
		}
		rpt.kinds = append(rpt.kinds, kr)
	}
	return rpt, nil
}

func gcLiveTokens() (map[string]bool, error) {
	trx, err := str.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "error starting transaction")
	}
	defer func() { _ = trx.Rollback() }()
	shrs, err := str.FindAllShares(trx)
	if err != nil {
		return nil, errors.Wrap(err, "error listing live shares")
	}
	live := make(map[string]bool, len(shrs))
	for _, shr := range shrs {
		live[shr.Token] = true
	}
	return live, nil
}

// gcReclaim deletes the survey's orphans by id, kind by kind in deletion order. an object already gone
// counts as collected; any other failure is logged and the run continues. it returns the failure count.
func gcReclaim(rpt *gcReport) int {
	failed := 0
	for _, kr := range rpt.kinds {
		for _, obj := range kr.orphans {
			err := kr.kind.delete(obj.id)
			switch {
			case err == nil:
				kr.deleted++
			case automation.IsNotFound(err):
				kr.alreadyGone++
				dl.Infof("%s '%s' ('%s') already gone", kr.kind.name, obj.id, obj.name)
			default:
				kr.failed++
				failed++
				dl.Errorf("error collecting %s '%s' ('%s'), share '%s': %v", kr.kind.name, obj.id, obj.name, obj.shareToken, err)
			}
		}
	}
	return failed
}

func (rpt *gcReport) print(out io.Writer) {
	if rpt.deleteMode {
		_, _ = fmt.Fprintf(out, "garbage collection (deleting orphans)\n")
	} else {
		_, _ = fmt.Fprintf(out, "garbage collection dry run (nothing deleted; run with --delete to remove orphans)\n")
	}
	_, _ = fmt.Fprintf(out, "live shares: %d, minimum age: %v\n\n", rpt.liveShares, rpt.minAge)

	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintf(tw, "kind\tlive\torphaned\ttoo young\tnot share-owned\t\n")
	for _, kr := range rpt.kinds {
		_, _ = fmt.Fprintf(tw, "%s\t%d\t%d\t%d\t%d\t\n", kr.kind.name, kr.live, len(kr.orphans), kr.tooYoung, kr.unowned)
	}
	_ = tw.Flush()

	byToken := make(map[string][]string)
	for _, kr := range rpt.kinds {
		for _, obj := range kr.orphans {
			byToken[obj.shareToken] = append(byToken[obj.shareToken], fmt.Sprintf("%s '%s' ('%s')", kr.kind.name, obj.id, obj.name))
		}
	}
	if len(byToken) == 0 {
		_, _ = fmt.Fprintf(out, "\nno orphaned objects\n")
		return
	}
	tokens := make([]string, 0, len(byToken))
	for token := range byToken {
		tokens = append(tokens, token)
	}
	sort.Strings(tokens)
	_, _ = fmt.Fprintf(out, "\norphaned objects by share token:\n")
	for _, token := range tokens {
		_, _ = fmt.Fprintf(out, "  '%s'\n", token)
		for _, line := range byToken[token] {
			_, _ = fmt.Fprintf(out, "    %s\n", line)
		}
	}
}

func (rpt *gcReport) printDeleted(out io.Writer) {
	_, _ = fmt.Fprintf(out, "\n")
	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintf(tw, "kind\tdeleted\talready gone\tfailed\t\n")
	for _, kr := range rpt.kinds {
		_, _ = fmt.Fprintf(tw, "%s\t%d\t%d\t%d\t\n", kr.kind.name, kr.deleted, kr.alreadyGone, kr.failed)
	}
	_ = tw.Flush()
}
