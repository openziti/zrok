package basicauth

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

// sentinel is a distinctive password; the no-quote tests assert it never appears in an error.
const sentinel = "zrokSentinelPw7Q"

func TestParseEntry(t *testing.T) {
	tests := map[string]struct {
		entry    string
		username string
		password string
	}{
		"plain":              {"admin:secret", "admin", "secret"},
		"colon in password":  {"admin:s3:cret", "admin", "s3:cret"},
		"trailing colon":     {"admin:secret:", "admin", "secret:"},
		"trimmed":            {"  admin \t: secret  ", "admin", "secret"},
		"empty password":     {"admin:", "admin", ""},
		"empty username":     {":secret", "", "secret"},
		"only a colon":       {":", "", ""},
		"colon-led password": {"admin::secret", "admin", ":secret"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			username, password, err := ParseEntry(tc.entry)
			require.NoError(t, err)
			require.Equal(t, tc.username, username)
			require.Equal(t, tc.password, password)
		})
	}
}

func TestParseEntryMalformedDoesNotQuote(t *testing.T) {
	for _, entry := range []string{sentinel, "", "  " + sentinel + "  "} {
		_, _, err := ParseEntry(entry)
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrMalformedEntry))
		require.NotContains(t, err.Error(), sentinel)
	}
}
