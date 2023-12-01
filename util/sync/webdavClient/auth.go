package webdavClient

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
)

// AuthFactory prototype function to create a new Authenticator
type AuthFactory func(c *http.Client, rs *http.Response, path string) (auth Authenticator, err error)

// Authorizer our Authenticator factory which creates an
// `Authenticator` per action/request.
type Authorizer interface {
	// Creates a new Authenticator Shim per request.
	// It may track request related states and perform payload buffering
	// for authentication round trips.
	// The underlying Authenticator will perform the real authentication.
	NewAuthenticator(body io.Reader) (Authenticator, io.Reader)
	// Registers a new Authenticator factory to a key.
	AddAuthenticator(key string, fn AuthFactory)
}

// A Authenticator implements a specific way to authorize requests.
// Each request is bound to a separate Authenticator instance.
//
// The authentication flow itself is broken down into `Authorize`
// and `Verify` steps. The former method runs before, and the latter
// runs after the `Request` is submitted.
// This makes it easy to encapsulate and control complex
// authentication challenges.
//
// Some authentication flows causing authentication round trips,
// which can be archived by returning the `redo` of the Verify
// method. `True` restarts the authentication process for the
// current action: A new `Request` is spawned, which must be
// authorized, sent, and re-verified again, until the action
// is successfully submitted.
// The preferred way is to handle the authentication ping-pong
// within `Verify`, and then `redo` with fresh credentials.
//
// The result of the `Verify` method can also trigger an
// `Authenticator` change by returning the `ErrAuthChanged`
// as an error. Depending on the `Authorizer` this may trigger
// an `Authenticator` negotiation.
//
// Set the `XInhibitRedirect` header to '1' in the `Authorize`
// method to get control over request redirection.
// Attention! You must handle the incoming request yourself.
//
// To store a shared session state the `Clone` method **must**
// return a new instance, initialized with the shared state.
type Authenticator interface {
	// Authorizes a request. Usually by adding some authorization headers.
	Authorize(c *http.Client, rq *http.Request, path string) error
	// Verifies the response if the authorization was successful.
	// May trigger some round trips to pass the authentication.
	// May also trigger a new Authenticator negotiation by returning `ErrAuthChenged`
	Verify(c *http.Client, rs *http.Response, path string) (redo bool, err error)
	// Creates a copy of the underlying Authenticator.
	Clone() Authenticator
	io.Closer
}

type authfactory struct {
	key    string
	create AuthFactory
}

// authorizer structure holds our Authenticator create functions
type authorizer struct {
	factories  []authfactory
	defAuthMux sync.Mutex
	defAuth    Authenticator
}

// preemptiveAuthorizer structure holds the preemptive Authenticator
type preemptiveAuthorizer struct {
	auth Authenticator
}

// authShim structure that wraps the real Authenticator
type authShim struct {
	factory AuthFactory
	body    io.Reader
	auth    Authenticator
}

// negoAuth structure holds the authenticators that are going to be negotiated
type negoAuth struct {
	auths                   []Authenticator
	setDefaultAuthenticator func(auth Authenticator)
}

// nullAuth initializes the whole authentication flow
type nullAuth struct{}

// noAuth structure to perform no authentication at all
type noAuth struct{}

// NewAutoAuth creates an auto Authenticator factory.
// It negotiates the default authentication method
// based on the order of the registered Authenticators
// and the remotely offered authentication methods.
// First In, First Out.
func NewAutoAuth(login string, secret string) Authorizer {
	fmap := make([]authfactory, 0)
	az := &authorizer{factories: fmap, defAuthMux: sync.Mutex{}, defAuth: &nullAuth{}}

	az.AddAuthenticator("basic", func(c *http.Client, rs *http.Response, path string) (auth Authenticator, err error) {
		return &BasicAuth{user: login, pw: secret}, nil
	})

	az.AddAuthenticator("digest", func(c *http.Client, rs *http.Response, path string) (auth Authenticator, err error) {
		return NewDigestAuth(login, secret, rs)
	})

	az.AddAuthenticator("passport1.4", func(c *http.Client, rs *http.Response, path string) (auth Authenticator, err error) {
		return NewPassportAuth(c, login, secret, rs.Request.URL.String(), &rs.Header)
	})

	return az
}

// NewEmptyAuth creates an empty Authenticator factory
// The order of adding the Authenticator matters.
// First In, First Out.
// It offers the `NewAutoAuth` features.
func NewEmptyAuth() Authorizer {
	fmap := make([]authfactory, 0)
	az := &authorizer{factories: fmap, defAuthMux: sync.Mutex{}, defAuth: &nullAuth{}}
	return az
}

// NewPreemptiveAuth creates a preemptive Authenticator
// The preemptive authorizer uses the provided Authenticator
// for every request regardless of any `Www-Authenticate` header.
//
// It may only have one authentication method,
// so calling `AddAuthenticator` **will panic**!
//
// Look out!! This offers the skinniest and slickest implementation
// without any synchronisation!!
// Still applicable with `BasicAuth` within go routines.
func NewPreemptiveAuth(auth Authenticator) Authorizer {
	return &preemptiveAuthorizer{auth: auth}
}

// NewAuthenticator creates an Authenticator (Shim) per request
func (a *authorizer) NewAuthenticator(body io.Reader) (Authenticator, io.Reader) {
	var retryBuf = body
	if body != nil {
		// If the authorization fails, we will need to restart reading
		// from the passed body stream.
		// When body is seekable, use seek to reset the streams
		// cursor to the start.
		// Otherwise, copy the stream into a buffer while uploading
		// and use the buffers content on retry.
		if _, ok := retryBuf.(io.Seeker); ok {
			body = io.NopCloser(body)
		} else {
			buff := &bytes.Buffer{}
			retryBuf = buff
			body = io.TeeReader(body, buff)
		}
	}
	a.defAuthMux.Lock()
	defAuth := a.defAuth.Clone()
	a.defAuthMux.Unlock()

	return &authShim{factory: a.factory, body: retryBuf, auth: defAuth}, body
}

// AddAuthenticator appends the AuthFactory to our factories.
// It converts the key to lower case and preserves the order.
func (a *authorizer) AddAuthenticator(key string, fn AuthFactory) {
	key = strings.ToLower(key)
	for _, f := range a.factories {
		if f.key == key {
			panic("Authenticator exists: " + key)
		}
	}
	a.factories = append(a.factories, authfactory{key, fn})
}

// factory picks all valid Authenticators based on Www-Authenticate headers
func (a *authorizer) factory(c *http.Client, rs *http.Response, path string) (auth Authenticator, err error) {
	headers := rs.Header.Values("Www-Authenticate")
	if len(headers) > 0 {
		auths := make([]Authenticator, 0)
		for _, f := range a.factories {
			for _, header := range headers {
				headerLower := strings.ToLower(header)
				if strings.Contains(headerLower, f.key) {
					rs.Header.Set("Www-Authenticate", header)
					if auth, err = f.create(c, rs, path); err == nil {
						auths = append(auths, auth)
						break
					}
				}
			}
		}

		switch len(auths) {
		case 0:
			return nil, NewPathError("NoAuthenticator", path, rs.StatusCode)
		case 1:
			auth = auths[0]
		default:
			auth = &negoAuth{auths: auths, setDefaultAuthenticator: a.setDefaultAuthenticator}
		}
	} else {
		auth = &noAuth{}
	}

	a.setDefaultAuthenticator(auth)

	return auth, nil
}

// setDefaultAuthenticator sets the default Authenticator
func (a *authorizer) setDefaultAuthenticator(auth Authenticator) {
	a.defAuthMux.Lock()
	a.defAuth.Close()
	a.defAuth = auth
	a.defAuthMux.Unlock()
}

// Authorize the current request
func (s *authShim) Authorize(c *http.Client, rq *http.Request, path string) error {
	if err := s.auth.Authorize(c, rq, path); err != nil {
		return err
	}
	body := s.body
	rq.GetBody = func() (io.ReadCloser, error) {
		if body != nil {
			if sk, ok := body.(io.Seeker); ok {
				if _, err := sk.Seek(0, io.SeekStart); err != nil {
					return nil, err
				}
			}
			return io.NopCloser(body), nil
		}
		return nil, nil
	}
	return nil
}

// Verify checks for authentication issues and may trigger a re-authentication.
// Catches AlgoChangedErr to update the current Authenticator
func (s *authShim) Verify(c *http.Client, rs *http.Response, path string) (redo bool, err error) {
	redo, err = s.auth.Verify(c, rs, path)
	if err != nil && errors.Is(err, ErrAuthChanged) {
		if auth, aerr := s.factory(c, rs, path); aerr == nil {
			s.auth.Close()
			s.auth = auth
			return true, nil
		} else {
			return false, aerr
		}
	}
	return
}

// Close closes all resources
func (s *authShim) Close() error {
	s.auth.Close()
	s.auth, s.factory = nil, nil
	if s.body != nil {
		if closer, ok := s.body.(io.Closer); ok {
			return closer.Close()
		}
	}
	return nil
}

// It's not intend to Clone the shim
// therefore it returns a noAuth instance
func (s *authShim) Clone() Authenticator {
	return &noAuth{}
}

// String toString
func (s *authShim) String() string {
	return "AuthShim"
}

// Authorize authorizes the current request with the top most Authorizer
func (n *negoAuth) Authorize(c *http.Client, rq *http.Request, path string) error {
	if len(n.auths) == 0 {
		return NewPathError("NoAuthenticator", path, 400)
	}
	return n.auths[0].Authorize(c, rq, path)
}

// Verify verifies the authentication and selects the next one based on the result
func (n *negoAuth) Verify(c *http.Client, rs *http.Response, path string) (redo bool, err error) {
	if len(n.auths) == 0 {
		return false, NewPathError("NoAuthenticator", path, 400)
	}
	redo, err = n.auths[0].Verify(c, rs, path)
	if err != nil {
		if len(n.auths) > 1 {
			n.auths[0].Close()
			n.auths = n.auths[1:]
			return true, nil
		}
	} else if redo {
		return
	} else {
		auth := n.auths[0]
		n.auths = n.auths[1:]
		n.setDefaultAuthenticator(auth)
		return
	}

	return false, NewPathError("NoAuthenticator", path, rs.StatusCode)
}

// Close will close the underlying authenticators.
func (n *negoAuth) Close() error {
	for _, a := range n.auths {
		a.Close()
	}
	n.setDefaultAuthenticator = nil
	return nil
}

// Clone clones the underlying authenticators.
func (n *negoAuth) Clone() Authenticator {
	auths := make([]Authenticator, len(n.auths))
	for i, e := range n.auths {
		auths[i] = e.Clone()
	}
	return &negoAuth{auths: auths, setDefaultAuthenticator: n.setDefaultAuthenticator}
}

func (n *negoAuth) String() string {
	return "NegoAuth"
}

// Authorize the current request
func (n *noAuth) Authorize(c *http.Client, rq *http.Request, path string) error {
	return nil
}

// Verify checks for authentication issues and may trigger a re-authentication
func (n *noAuth) Verify(c *http.Client, rs *http.Response, path string) (redo bool, err error) {
	if "" != rs.Header.Get("Www-Authenticate") {
		err = ErrAuthChanged
	}
	return
}

// Close closes all resources
func (n *noAuth) Close() error {
	return nil
}

// Clone creates a copy of itself
func (n *noAuth) Clone() Authenticator {
	// no copy due to read only access
	return n
}

// String toString
func (n *noAuth) String() string {
	return "NoAuth"
}

// Authorize the current request
func (n *nullAuth) Authorize(c *http.Client, rq *http.Request, path string) error {
	rq.Header.Set(XInhibitRedirect, "1")
	return nil
}

// Verify checks for authentication issues and may trigger a re-authentication
func (n *nullAuth) Verify(c *http.Client, rs *http.Response, path string) (redo bool, err error) {
	return true, ErrAuthChanged
}

// Close closes all resources
func (n *nullAuth) Close() error {
	return nil
}

// Clone creates a copy of itself
func (n *nullAuth) Clone() Authenticator {
	// no copy due to read only access
	return n
}

// String toString
func (n *nullAuth) String() string {
	return "NullAuth"
}

// NewAuthenticator creates an Authenticator (Shim) per request
func (b *preemptiveAuthorizer) NewAuthenticator(body io.Reader) (Authenticator, io.Reader) {
	return b.auth.Clone(), body
}

// AddAuthenticator Will PANIC because it may only have a single authentication method
func (b *preemptiveAuthorizer) AddAuthenticator(key string, fn AuthFactory) {
	panic("You're funny! A preemptive authorizer may only have a single authentication method")
}
