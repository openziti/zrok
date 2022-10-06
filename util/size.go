package util

import "fmt"

func BytesToSize(sz int64) string {
	absSz := sz
	if absSz < 0 {
		absSz *= -1
	}

	const unit = 1000
	if absSz < unit {
		return fmt.Sprintf("%d B", sz)
	}
	div, exp := int64(unit), 0
	for n := absSz / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(sz)/float64(div), "kMGTPE"[exp])
}
