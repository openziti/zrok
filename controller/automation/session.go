package automation

import (
	"context"
	"crypto/x509"
	"expvar"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/edge-api/rest_management_api_client"
	"github.com/openziti/edge-api/rest_util"
)

var zitiAuthentications = expvar.NewInt("zrok.ziti.authentications")

// failedRefreshWindow is how long a failed refresh answers callers still holding the same session
// generation, so a persistent authentication failure costs one login per window.
const failedRefreshWindow = 5 * time.Second

// refreshClock is replaceable so tests can move past failedRefreshWindow.
var refreshClock = time.Now

type sessionKey struct {
	endpoint, username string
}

type sessionCache struct {
	mu       sync.Mutex
	clients  map[sessionKey]*ZitiAutomation
	building map[sessionKey]*sessionBuild
	caPool   func(string) (*x509.CertPool, error)
}

type sessionBuild struct {
	done chan struct{}
	ziti *ZitiAutomation
	err  error
}

var sharedSessions = sessionCache{
	clients:  make(map[sessionKey]*ZitiAutomation),
	building: make(map[sessionKey]*sessionBuild),
	caPool:   rest_util.GetControllerWellKnownCaPool,
}

func (c *sessionCache) get(cfg *Config) (*ZitiAutomation, error) {
	key := sessionKey{cfg.ApiEndpoint, cfg.Username}
	c.mu.Lock()
	if ziti := c.clients[key]; ziti != nil {
		c.mu.Unlock()
		return ziti, nil
	}
	if build := c.building[key]; build != nil {
		c.mu.Unlock()
		<-build.done
		return build.ziti, build.err
	}
	// the first build runs outside the lock so a slow login blocks only callers for this key.
	build := &sessionBuild{done: make(chan struct{})}
	c.building[key] = build
	caPool := c.caPool
	c.mu.Unlock()

	pool, err := caPool(cfg.ApiEndpoint)
	var ziti *ZitiAutomation
	if err == nil {
		ziti, err = newZitiSession(cfg, pool)
	}

	c.mu.Lock()
	delete(c.building, key)
	if err == nil {
		c.clients[key] = ziti
	}
	c.mu.Unlock()
	build.ziti, build.err = ziti, err
	close(build.done)
	return ziti, err
}

// newZitiSession accepts an explicit CA pool so tests can use a plain HTTP fake.
func newZitiSession(cfg *Config, pool *x509.CertPool) (*ZitiAutomation, error) {
	endpoint, err := url.Parse(cfg.ApiEndpoint)
	if err != nil {
		return nil, err
	}
	if endpoint.Host == "" || (endpoint.Scheme != "https" && endpoint.Scheme != "http") {
		return nil, fmt.Errorf("invalid ziti api endpoint '%s'", cfg.ApiEndpoint)
	}
	auth := rest_util.NewAuthenticatorUpdb(cfg.Username, cfg.Password)
	auth.RootCas = pool
	client, err := auth.BuildHttpClient()
	if err != nil {
		return nil, err
	}
	basePath := endpoint.Path
	if basePath == "" || basePath == "/" {
		basePath = rest_management_api_client.DefaultBasePath
	}
	newEdge := func(client *http.Client) *rest_management_api_client.ZitiEdgeManagement {
		runtime := httptransport.NewWithClient(endpoint.Host, basePath, []string{endpoint.Scheme}, client)
		return rest_management_api_client.New(runtime, nil)
	}
	// authentication bypasses the session transport and uses the same TLS pool.
	authEdge := newEdge(client)
	session := &sessionTransport{base: client.Transport}
	session.authenticate = func(ctx context.Context) (string, error) {
		params := auth.Params().WithContext(ctx).WithTimeout(DefaultRequestTimeout)
		resp, err := authEdge.Authentication.Authenticate(params)
		if err != nil {
			return "", err
		}
		if resp.Payload == nil || resp.Payload.Data == nil || resp.Payload.Data.Token == nil || *resp.Payload.Data.Token == "" {
			return "", fmt.Errorf("ziti api session token was empty")
		}
		return *resp.Payload.Data.Token, nil
	}
	if err := session.refresh(context.Background(), 0, "initial"); err != nil {
		client.CloseIdleConnections()
		return nil, err
	}
	sessionClient := *client
	sessionClient.Transport = session
	return NewZitiAutomationWithEdge(newEdge(&sessionClient)), nil
}

type sessionTransport struct {
	base         http.RoundTripper
	authenticate func(context.Context) (string, error)
	mu           sync.RWMutex
	token        string
	generation   uint64
	inflight     *sessionRefresh
	failure      *sessionRefreshFailure
}

type sessionRefresh struct {
	done chan struct{}
	err  error
}

type sessionRefreshFailure struct {
	err        error
	generation uint64
	at         time.Time
}

// refresh is single-flight: concurrent callers share one authentication attempt, and the lock is
// never held across the network call.
func (s *sessionTransport) refresh(ctx context.Context, generation uint64, reason string) error {
	s.mu.Lock()
	if s.generation != generation {
		s.mu.Unlock()
		return nil
	}
	if f := s.failure; f != nil && f.generation == generation && refreshClock().Sub(f.at) < failedRefreshWindow {
		s.mu.Unlock()
		return f.err
	}
	call := s.inflight
	if call == nil {
		call = &sessionRefresh{done: make(chan struct{})}
		s.inflight = call
		// detached from the caller's cancellation so one abandoned request cannot fail every waiter.
		go s.runRefresh(context.WithoutCancel(ctx), call, generation, reason)
	}
	s.mu.Unlock()
	select {
	case <-call.done:
		return call.err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *sessionTransport) runRefresh(ctx context.Context, call *sessionRefresh, generation uint64, reason string) {
	token, err := s.authenticate(ctx)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.inflight = nil
	if err != nil {
		s.failure = &sessionRefreshFailure{err: err, generation: generation, at: refreshClock()}
	} else {
		s.token = token
		s.generation++
		s.failure = nil
		zitiAuthentications.Add(1)
		dl.Infof("ziti authentication count '%d', reason '%s'", zitiAuthentications.Value(), reason)
	}
	call.err = err
	close(call.done)
}

func (s *sessionTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	request := req.Clone(req.Context())
	s.mu.RLock()
	request.Header.Set("zt-session", s.token)
	generation := s.generation
	s.mu.RUnlock()
	resp, err := s.base.RoundTrip(request)
	if err != nil || resp.StatusCode != http.StatusUnauthorized {
		return resp, err
	}
	// a streaming body cannot be replayed; preserve the original typed 401.
	if req.Body != nil && req.Body != http.NoBody && req.GetBody == nil {
		return resp, nil
	}
	if err := s.refresh(req.Context(), generation, "expired"); err != nil {
		// let the generated client decode the original operation's typed 401.
		return resp, nil
	}
	resp.Body.Close()
	replay := req.Clone(req.Context())
	if req.GetBody != nil {
		replay.Body, err = req.GetBody()
		if err != nil {
			return nil, err
		}
	}
	s.mu.RLock()
	replay.Header.Set("zt-session", s.token)
	s.mu.RUnlock()
	return s.base.RoundTrip(replay)
}
