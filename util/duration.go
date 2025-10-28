package util

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// ParseDuration extends time.ParseDuration to support 'd' (days) as a unit.
// it converts days to hours (1d = 24h) before parsing.
// examples: "24h", "7d", "2d6h30m", "1d12h"
func ParseDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, fmt.Errorf("invalid duration: empty string")
	}

	// check for invalid patterns that might contain 'd' but shouldn't be processed
	// this catches cases like "-1d", "1.5d", ".5d" etc.
	invalidPattern := regexp.MustCompile(`[^\d\s](\d+)d|(\d*\.\d+)d`)
	if invalidPattern.MatchString(s) {
		// let time.ParseDuration handle these and return its error
		return time.ParseDuration(s)
	}

	// regex pattern to match digits followed by 'd' at word boundaries
	dayPattern := regexp.MustCompile(`(\d+)d`)

	// find all day values and convert to hours
	converted := dayPattern.ReplaceAllStringFunc(s, func(match string) string {
		// extract the numeric part
		numStr := match[:len(match)-1] // remove the 'd'
		days, err := strconv.Atoi(numStr)
		if err != nil {
			// this shouldn't happen due to regex, but handle gracefully
			return match
		}
		// convert days to hours
		hours := days * 24
		return fmt.Sprintf("%dh", hours)
	})

	// pass to standard time.ParseDuration
	return time.ParseDuration(converted)
}
