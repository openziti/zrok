package limits

import (
	"github.com/openziti/zrok/controller/store"
	"sort"
)

func sortLimitClasses(lcs []*store.LimitClass) {
	sort.Slice(lcs, func(i, j int) bool {
		return modePoints(lcs[i]) > modePoints(lcs[j])
	})
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
