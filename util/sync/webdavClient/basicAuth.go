package webdavClient

import (
	"fmt"
	"net/http"
)

// BasicAuth structure holds our credentials
type BasicAuth struct {
	user string
	pw   string
}

// Authorize the current request
func (b *BasicAuth) Authorize(c *http.Client, rq *http.Request, path string) error {
	rq.SetBasicAuth(b.user, b.pw)
	return nil
}

// Verify verifies if the authentication
func (b *BasicAuth) Verify(c *http.Client, rs *http.Response, path string) (redo bool, err error) {
	if rs.StatusCode == 401 {
		err = NewPathError("Authorize", path, rs.StatusCode)
	}
	return
}

// Close cleans up all resources
func (b *BasicAuth) Close() error {
	return nil
}

// Clone creates a Copy of itself
func (b *BasicAuth) Clone() Authenticator {
	// no copy due to read only access
	return b
}

// String toString
func (b *BasicAuth) String() string {
	return fmt.Sprintf("BasicAuth login: %s", b.user)
}
