package limits

import (
	"github.com/openziti/zrok/controller/store"
	"sort"
)

func sortLimitClasses(lcs []*store.LimitClass) {
	sort.Slice(lcs, func(i, j int) bool {
		ipoints := limitScopePoints(lcs[i]) + modePoints(lcs[i])
		jpoints := limitScopePoints(lcs[j]) + modePoints(lcs[j])
		return ipoints > jpoints
	})
}

func limitScopePoints(lc *store.LimitClass) int {
	points := 0
	switch lc.LimitScope {
	case store.AccountLimitScope:
		points += 1000
	case store.EnvironmentLimitScope:
		points += 100
	case store.ShareLimitScope:
		points += 10
	}
	return points
}

func modePoints(lc *store.LimitClass) int {
	points := 0
	if lc.BackendMode != "" {
		points += 1
	}
	if lc.ShareMode != "" {
		points += 1
	}
	return points
}
