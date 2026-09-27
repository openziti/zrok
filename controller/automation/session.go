package automation

import (
	"context"
	"crypto/x509"
	"expvar"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/edge-api/rest_management_api_client"
	"github.com/openziti/edge-api/rest_util"
)

var zitiAuthentications = expvar.NewInt("zrok.ziti.authentications")

type sessionKey struct {
	endpoint, username string
}

type sessionCache struct {
	mu      sync.Mutex
	clients map[sessionKey]*ZitiAutomation
	caPool  func(string) (*x509.CertPool, error)
}

var sharedSessions = sessionCache{
	clients: make(map[sessionKey]*ZitiAutomation),
	caPool:  rest_util.GetControllerWellKnownCaPool,
}

func (c *sessionCache) get(cfg *Config) (*ZitiAutomation, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := sessionKey{cfg.ApiEndpoint, cfg.Username}
	if ziti := c.clients[key]; ziti != nil {
		return ziti, nil
	}
	pool, err := c.caPool(cfg.ApiEndpoint)
	if err != nil {
		return nil, err
	}
	ziti, err := newZitiSession(cfg, pool)
	if err != nil {
		return nil, err
	}
	c.clients[key] = ziti
	return ziti, nil
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
}

func (s *sessionTransport) refresh(ctx context.Context, generation uint64, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.generation != generation {
		return nil
	}
	token, err := s.authenticate(ctx)
	if err != nil {
		return err
	}
	s.token = token
	s.generation++
	zitiAuthentications.Add(1)
	dl.Infof("ziti authentication count '%d', reason '%s'", zitiAuthentications.Value(), reason)
	return nil
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
