package dynamicProxy

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/michaelquigley/df/da"
	"github.com/michaelquigley/df/dd"
	"github.com/openziti/zrok/endpoints/proxyUi"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var oauthBinders map[string]func(map[string]any) (dd.Dynamic, error)

func registerOauthBinder(typ string, binder func(map[string]any) (dd.Dynamic, error)) {
	if oauthBinders == nil {
		oauthBinders = make(map[string]func(map[string]any) (dd.Dynamic, error))
	}
	oauthBinders[typ] = binder
}

var globalOAuthRouter *oauthRouter

func registerOAuthProvider(provider oauthProvider) error {
	if globalOAuthRouter == nil {
		return errors.New("oauth router not initialized")
	}
	return globalOAuthRouter.registerProvider(provider)
}

// oauthRouter manages OAuth provider routes
type oauthRouter struct {
	cfg       *oauthConfig
	router    *mux.Router
	server    *http.Server
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.RWMutex
	providers map[string]oauthProvider
}

// oauthProvider interface for OAuth providers that can register their own routes
type oauthProvider interface {
	Name() string
	RegisterRoutes(router *mux.Router) error
}

// newOAuthRouter creates a new OAuth router with the given configuration
func newOAuthRouter(cfg *oauthConfig) *oauthRouter {
	ctx, cancel := context.WithCancel(context.Background())
	router := mux.NewRouter()

	return &oauthRouter{
		cfg:       cfg,
		router:    router,
		ctx:       ctx,
		cancel:    cancel,
		providers: make(map[string]oauthProvider),
	}
}

// registerProvider registers an OAuth provider with the router
func (r *oauthRouter) registerProvider(provider oauthProvider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.providers[provider.Name()]; exists {
		return errors.Errorf("provider '%s' already registered", provider.Name())
	}

	if err := provider.RegisterRoutes(r.router); err != nil {
		return errors.Wrapf(err, "failed to register routes for provider '%s'", provider.Name())
	}

	r.providers[provider.Name()] = provider
	logrus.Debugf("registered oauth provider: '%s'", provider.Name())

	return nil
}

// Start starts the OAuth HTTP server
func (r *oauthRouter) Start() error {
	// set up default route for unauthorized requests
	r.router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		proxyUi.WriteUnauthorized(w, proxyUi.UnauthorizedData())
	})

	r.server = &http.Server{
		Addr:         r.cfg.BindAddress,
		Handler:      r.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logrus.Infof("starting oauth server on '%s'", r.cfg.BindAddress)
		if err := r.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.Errorf("oauth server error: %v", err)
		}
	}()

	logrus.Infof("started with '%d' providers", len(r.providers))

	return nil
}

func (r *oauthRouter) Stop() error {
	logrus.Info("stopping")
	if r.cancel != nil {
		r.cancel()
	}
	if r.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return r.server.Shutdown(ctx)
	}
	return nil
}

// buildOAuthRouter creates and registers the OAuth router as a df component
func buildOAuthRouter(app *da.Application[*config]) error {
	if app.Cfg.Oauth == nil {
		logrus.Info("no oauth configuration; skipping oauth router")
		return nil
	}

	// initialize oauth router as global (still needed for provider registration)
	globalOAuthRouter = newOAuthRouter(app.Cfg.Oauth)

	// register with df container for lifecycle management
	da.Set(app.C, globalOAuthRouter)
	return nil
}
