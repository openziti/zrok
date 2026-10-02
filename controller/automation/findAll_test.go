package automation

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestFindAllClampsPageSize lists through a finder that caps every page at MaxPageSize, as ziti does; a
// requested page above the cap must not read the capped answer as the last page.
func TestFindAllClampsPageSize(t *testing.T) {
	items := make([]*int, 1203)
	for i := range items {
		items[i] = new(int)
		*items[i] = i
	}
	var limits []int64
	finder := func(opts *FilterOptions) ([]*int, error) {
		limits = append(limits, opts.Limit)
		limit := min(opts.Limit, MaxPageSize)
		start := min(opts.Offset, int64(len(items)))
		return items[start:min(start+limit, int64(len(items)))], nil
	}
	for _, pageSize := range []int64{0, MaxPageSize + 500} {
		limits = nil
		all, err := FindAll(finder, "", pageSize)
		require.NoError(t, err)
		require.Equal(t, items, all)
		require.Equal(t, []int64{MaxPageSize, MaxPageSize, MaxPageSize}, limits)
	}
}
