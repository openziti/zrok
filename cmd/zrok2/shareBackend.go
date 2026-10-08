package main

import (
	"sync"

	"github.com/michaelquigley/df/dl"
)

// shareBackend is what the share commands drive of a backend.
type shareBackend interface {
	Run() error
	Stop() error
}

// startsInline reports whether a backend mode's Run returns once the backend has started, rather than
// serving until it is stopped. the caddy backends hand their configuration to caddy and return, so an
// error from their Run is a failure to start.
func startsInline(backendMode string) bool {
	return backendMode == "web" || backendMode == "caddy"
}

// runningBackend holds a share's backend once it is created, so a shutdown can stop it before the share is
// deleted.
type runningBackend struct {
	mu       sync.Mutex
	be       shareBackend
	stopping bool
}

func (r *runningBackend) set(be shareBackend) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.be = be
}

// stop closes the backend's listener, if a backend was created; a Run returning after stop is not a
// failure.
func (r *runningBackend) stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stopping = true
	if r.be != nil {
		if err := r.be.Stop(); err != nil {
			dl.Warnf("error stopping backend: %v", err)
		}
		r.be = nil
	}
}

func (r *runningBackend) isStopping() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.stopping
}
