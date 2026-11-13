// package basicauth is the one grammar for a basic authentication entry. it imports nothing from the module so
// the sdk and share preparation can both use it. no error from this package carries any part of an entry.
package basicauth

import (
	"errors"
	"strings"
)

// ErrMalformedEntry reports an entry with no colon separating its username from its password.
var ErrMalformedEntry = errors.New("expected <username:password>")

// ParseEntry splits entry at its first colon and trims the username and the password. a username cannot
// contain a colon; a password can. this is the split the frontends apply to an Authorization header.
func ParseEntry(entry string) (username, password string, err error) {
	username, password, found := strings.Cut(entry, ":")
	if !found {
		return "", "", ErrMalformedEntry
	}
	return strings.TrimSpace(username), strings.TrimSpace(password), nil
}
