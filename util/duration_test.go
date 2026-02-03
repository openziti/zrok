package util

import (
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
		wantErr  bool
	}{
		// standard formats (pass-through)
		{"standard hours", "24h", 24 * time.Hour, false},
		{"standard minutes", "90m", 90 * time.Minute, false},
		{"standard combined", "1h30m", 90 * time.Minute, false},
		{"standard seconds", "45s", 45 * time.Second, false},

		// days only
		{"one day", "1d", 24 * time.Hour, false},
		{"seven days", "7d", 168 * time.Hour, false},
		{"thirty days", "30d", 720 * time.Hour, false},

		// combined formats
		{"days and hours", "1d12h", 36 * time.Hour, false},
		{"days hours minutes", "2d6h30m", 54*time.Hour + 30*time.Minute, false},
		{"complex", "3d2h15m30s", 74*time.Hour + 15*time.Minute + 30*time.Second, false},

		// edge cases
		{"zero days", "0d", 0, false},
		{"empty string", "", 0, true},
		{"invalid format", "invalid", 0, true},
		{"days without number", "d", 0, true},
		{"negative days", "-1d", 0, true},    // '-' not in \d+ pattern, passes through as invalid
		{"decimal days", "1.5d", 0, true}, // '.' not in \d+ pattern, passes through as invalid
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDuration(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseDuration(%q) expected error but got none", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseDuration(%q) unexpected error: %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("ParseDuration(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}
