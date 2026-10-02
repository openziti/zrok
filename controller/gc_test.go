package controller

import (
	"bytes"
	"fmt"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/openziti/zrok/v2/controller/automation"
	"github.com/openziti/zrok/v2/controller/automation/zitifake"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/stretchr/testify/require"
)

// gcKindOrder is the fake's name for each kind, in gc's deletion order.
var gcKindOrder = []string{zitifake.ServiceEdgeRouterPolicies, zitifake.ServicePolicies, zitifake.Configs, zitifake.Services}

const (
	gcLivePublic  = "livepub"
	gcLivePrivate = "livepriv"
	gcOrphan      = "orphan"
	gcYoungOrphan = "youngorphan"
)

type gcFixture struct {
	*shareCreateFixture
	ziti *automation.ZitiAutomation
	// orphans and kept are 'kind/id' for the old orphan's objects and for everything gc must leave.
	orphans, kept []string
	youngOrphans  []string
}

// setupGCFixture seeds the fake with objects tagged the way the real paths tag them: a live public share,
// a live private share with an access-path dial policy, an agent-remote set, an orphaned share whose row
// is deleted, an untagged object, a zrok-only object and a young orphan. everything but the young orphan
// is backdated past the age guard, so only the share token keeps the live shares' objects.
func setupGCFixture(t *testing.T) *gcFixture {
	t.Helper()
	f := &gcFixture{shareCreateFixture: setupShareCreateFixture(t)}
	var err error
	f.ziti, err = automation.NewZitiAutomation(cfg.Ziti)
	require.NoError(t, err)

	trx, err := str.Begin()
	require.NoError(t, err)
	envs, err := str.FindEnvironmentsForAccount(int(f.principal.ID), trx)
	require.NoError(t, err)
	envID := envs[0].Id
	for _, token := range []string{gcLivePublic, gcLivePrivate, gcOrphan} {
		mode := "public"
		if token == gcLivePrivate {
			mode = "private"
		}
		shrID, err := str.CreateShare(envID, &store.Share{ZId: token + "-svc", Token: token, ShareMode: mode, BackendMode: "proxy", PermissionMode: store.OpenPermissionMode}, trx)
		require.NoError(t, err)
		if token == gcOrphan {
			require.NoError(t, str.DeleteShare(shrID, trx))
		}
	}
	require.NoError(t, trx.Commit())

	old := 2 * DefaultGCMinAge
	seed := func(kind, id, name string, tags *automation.Tags, age time.Duration) string {
		if tags == nil {
			f.fake.SeedWithID(kind, id, name, nil)
		} else {
			f.fake.SeedWithID(kind, id, name, tags.ToRestModel())
		}
		if age > 0 {
			f.fake.Backdate(kind, id, age)
		}
		return kind + "/" + id
	}
	shareSet := func(token string, age time.Duration, dial bool) []string {
		tags := automation.ZrokShareTags(token)
		set := []string{
			seed(zitifake.Configs, token+"-cfg", token, tags, age),
			seed(zitifake.Services, token+"-svc", token, tags, age),
			seed(zitifake.ServicePolicies, token+"-bind", "env-zid-"+token+"-svc-bind", tags, age),
			seed(zitifake.ServiceEdgeRouterPolicies, token+"-serp", "env-zid-"+token+"-serp", tags, age),
		}
		if dial {
			set = append(set, seed(zitifake.ServicePolicies, token+"-dial", "env-zid-"+token+"-svc-dial", tags, age))
		}
		return set
	}

	f.kept = append(f.kept, shareSet(gcLivePublic, old, true)...)
	f.kept = append(f.kept, shareSet(gcLivePrivate, old, false)...)
	accessTags := automation.ZrokShareTags(gcLivePrivate).WithTag("zrokEnvironmentZId", "access-zid").WithTag("zrokFrontendToken", "fetoken")
	f.kept = append(f.kept, seed(zitifake.ServicePolicies, "access-dial", "fetoken-access-zid-"+gcLivePrivate+"-svc-dial", accessTags, old))

	remoteTags := automation.ZrokAgentRemoteTags("enrolltoken", "env-zid")
	f.kept = append(f.kept,
		seed(zitifake.Services, "remote-svc", "enrolltoken", remoteTags, old),
		seed(zitifake.ServicePolicies, "remote-bind", "env-zid-enrolltoken-bind", remoteTags, old),
		seed(zitifake.ServicePolicies, "remote-dial", "env-zid-enrolltoken-dial", remoteTags, old),
		seed(zitifake.ServiceEdgeRouterPolicies, "remote-serp", "enrolltoken", remoteTags, old),
	)

	f.kept = append(f.kept,
		seed(zitifake.Services, "untagged-svc", "untagged", nil, old),
		seed(zitifake.ServicePolicies, "zrok-only-policy", "zrok-only", automation.ZrokTags(), old),
	)

	f.orphans = shareSet(gcOrphan, old, true)
	f.youngOrphans = shareSet(gcYoungOrphan, 0, true)
	f.kept = append(f.kept, f.youngOrphans...)
	return f
}

func (f *gcFixture) run(t *testing.T, opts GCOptions) (*gcReport, string) {
	t.Helper()
	out := &bytes.Buffer{}
	rpt, err := gcSurvey(f.ziti, opts)
	require.NoError(t, err)
	rpt.print(out)
	if opts.Delete {
		gcReclaim(rpt)
		rpt.printDeleted(out)
	}
	return rpt, out.String()
}

// orphanSet is 'kind/id' for every orphan in the report, sorted.
func orphanSet(rpt *gcReport) []string {
	var out []string
	for i, kr := range rpt.kinds {
		for _, obj := range kr.orphans {
			out = append(out, gcKindOrder[i]+"/"+obj.id)
		}
	}
	sort.Strings(out)
	return out
}

func sorted(in []string) []string {
	out := slices.Clone(in)
	sort.Strings(out)
	return out
}

func gcDeletes(fake *zitifake.Server) int {
	_, deleted := fake.Log()
	return len(deleted)
}

func TestGCDryRunDeletesNothing(t *testing.T) {
	f := setupGCFixture(t)
	rpt, out := f.run(t, GCOptions{MinAge: DefaultGCMinAge})

	require.Zero(t, gcDeletes(f.fake))
	require.Equal(t, sorted(f.orphans), orphanSet(rpt))

	tooYoung, unowned, live := 0, 0, 0
	for _, kr := range rpt.kinds {
		tooYoung += kr.tooYoung
		unowned += kr.unowned
		live += kr.live
	}
	require.Equal(t, len(f.youngOrphans), tooYoung)
	// the agent-remote set and the zrok-only policy; the untagged service is never listed.
	require.Equal(t, 5, unowned)
	require.Equal(t, 10, live)

	require.Contains(t, out, "dry run")
	require.Contains(t, out, "'"+gcOrphan+"'")
	for _, token := range []string{gcLivePublic, gcLivePrivate, gcYoungOrphan, "enrolltoken", "untagged", "zrok-only"} {
		require.NotContains(t, out, token)
	}
}

func TestGCDeleteRemovesOnlyOrphansInOrder(t *testing.T) {
	f := setupGCFixture(t)
	rpt, out := f.run(t, GCOptions{Delete: true, MinAge: DefaultGCMinAge})

	_, deleted := f.fake.Log()
	require.Equal(t, sorted(f.orphans), sorted(deleted))
	rank := func(entry string) int {
		kind, _, _ := strings.Cut(entry, "/")
		return slices.Index(gcKindOrder, kind)
	}
	require.True(t, slices.IsSortedFunc(deleted, func(a, b string) int { return rank(a) - rank(b) }), "%v", deleted)

	for _, entry := range f.orphans {
		kind, id, _ := strings.Cut(entry, "/")
		require.False(t, f.fake.Has(kind, id), entry)
	}
	for _, entry := range f.kept {
		kind, id, _ := strings.Cut(entry, "/")
		require.True(t, f.fake.Has(kind, id), entry)
	}
	for _, kr := range rpt.kinds {
		require.Equal(t, len(kr.orphans), kr.deleted, kr.kind.name)
		require.Zero(t, kr.alreadyGone+kr.failed, kr.kind.name)
	}
	require.Contains(t, out, "already gone")
}

func TestGCReadsEveryPage(t *testing.T) {
	f := setupGCFixture(t)
	// more objects of each kind than three pages of the test's page size, and more than ziti's default
	// page of ten.
	old := 2 * DefaultGCMinAge
	for i := 0; i < 23; i++ {
		token := fmt.Sprintf("paged%02d", i)
		for _, kind := range gcKindOrder {
			id := fmt.Sprintf("%s-%s", token, kind)
			f.fake.SeedWithID(kind, id, id, automation.ZrokShareTags(token).ToRestModel())
			f.fake.Backdate(kind, id, old)
		}
	}

	rpt, _ := f.run(t, GCOptions{MinAge: DefaultGCMinAge, pageSize: 7})
	for i, kr := range rpt.kinds {
		// every object of the kind but the untagged service is zrok tagged and so examined.
		want := f.fake.Len(gcKindOrder[i])
		if gcKindOrder[i] == zitifake.Services {
			want--
		}
		require.Equal(t, want, kr.total(), kr.kind.name)
	}
	require.Len(t, orphanSet(rpt), len(f.orphans)+23*len(gcKindOrder))
	require.Zero(t, gcDeletes(f.fake))
}

func TestGCObjectGoneBeforeDeleteIsCollected(t *testing.T) {
	f := setupGCFixture(t)
	rpt, err := gcSurvey(f.ziti, GCOptions{Delete: true, MinAge: DefaultGCMinAge})
	require.NoError(t, err)

	// another process removes one orphaned service between the listing and its delete.
	f.fake.Remove(zitifake.Services, gcOrphan+"-svc")
	require.Zero(t, gcReclaim(rpt))
	out := &bytes.Buffer{}
	rpt.printDeleted(out)

	for _, entry := range f.orphans {
		kind, id, _ := strings.Cut(entry, "/")
		require.False(t, f.fake.Has(kind, id), entry)
	}
	services := rpt.kinds[slices.Index(gcKindOrder, zitifake.Services)]
	require.Equal(t, 1, services.alreadyGone)
	require.Zero(t, services.deleted)
	require.Zero(t, services.failed)
	_, deleted := f.fake.Log()
	require.Len(t, deleted, len(f.orphans)-1)
}

func TestGCDeleteFailureContinues(t *testing.T) {
	f := setupGCFixture(t)
	f.fake.RejectDeletes(zitifake.ServicePolicies, true)

	out := &bytes.Buffer{}
	err := runGC(f.ziti, GCOptions{Delete: true, MinAge: DefaultGCMinAge}, out)
	require.ErrorContains(t, err, "'2' deletes failed")

	for _, entry := range f.orphans {
		kind, id, _ := strings.Cut(entry, "/")
		require.Equal(t, kind == zitifake.ServicePolicies, f.fake.Has(kind, id), entry)
	}
	require.Contains(t, f.logs.String(), "error collecting service policies '"+gcOrphan+"-bind'")
	require.Contains(t, out.String(), "failed")
}
