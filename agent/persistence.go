package agent

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
)

func (a *Agent) ensureImportLocked() *registryLoad {
	if a.registryLoad != nil {
		return a.registryLoad
	}
	load := &registryLoad{done: make(chan struct{})}
	a.registryLoad = load
	a.beginWorkLocked()
	go func() {
		defer a.endWork()
		path, err := a.root.AgentRegistry()
		var registry *Registry
		if err == nil {
			registry, err = a.deps.loadRegistry(path)
		}
		if errors.Is(err, os.ErrNotExist) {
			registry = &Registry{V: RegistryV}
			err = nil
		}
		if err == nil {
			if registry == nil {
				err = errors.New("registry loader returned no registry")
			} else {
				for _, entry := range registry.PublicShares {
					if entry == nil || entry.Request == nil {
						err = errors.New("missing public share registry request")
					}
				}
				for _, entry := range registry.PrivateShares {
					if entry == nil || entry.Request == nil {
						err = errors.New("missing private share registry request")
					}
				}
				for _, entry := range registry.PrivateAccesses {
					if entry == nil || entry.Request == nil {
						err = errors.New("missing private access registry request")
					}
				}
			}
		}
		a.mu.Lock()
		if err == nil {
			importRequest := func(request any, failure *FailureEntry) {
				for _, existing := range a.operations {
					if existing.desired && sameRequest(existing.request, request) {
						return
					}
				}
				op := &operation{id: operationID(), request: cloneRequest(request), origin: "registry", phase: phaseQueued, desired: true}
				if failure != nil {
					copy := *failure
					op.failure = &copy
					op.failureID = op.id
				}
				a.operations[op.id] = op
			}
			for _, entry := range registry.PrivateAccesses {
				importRequest(entry.Request, entry.Failure)
			}
			for _, entry := range registry.PublicShares {
				importRequest(entry.Request, entry.Failure)
			}
			for _, entry := range registry.PrivateShares {
				importRequest(entry.Request, entry.Failure)
			}
		}
		load.err = err
		close(load.done)
		a.mu.Unlock()
	}()
	return load
}

func (a *Agent) imported() error {
	a.mu.Lock()
	load := a.ensureImportLocked()
	a.mu.Unlock()
	<-load.done
	return load.err
}

// waitForImport gates caller mutations on the imported snapshot, independently of worker replay.
func (a *Agent) waitForImport(ctx context.Context) error {
	a.mu.Lock()
	if a.stopping {
		a.mu.Unlock()
		return errors.New("agent is stopping")
	}
	if err := ctx.Err(); err != nil {
		a.mu.Unlock()
		return err
	}
	load := a.ensureImportLocked()
	a.mu.Unlock()
	select {
	case <-load.done:
		return load.err
	case <-ctx.Done():
		return ctx.Err()
	case <-a.ctx.Done():
		return errors.New("agent is stopping")
	}
}

func (a *Agent) RestoreRegistryOnce() error {
	if err := a.imported(); err != nil {
		return err
	}
	a.mu.Lock()
	if a.stopping {
		a.mu.Unlock()
		return nil
	}
	if a.replayStarted {
		a.mu.Unlock()
		return nil
	}
	a.replayStarted = true
	a.beginWorkLocked()
	var queued []*operation
	for _, op := range a.operations {
		if op.phase == phaseQueued && op.origin == "registry" {
			queued = append(queued, op)
		}
	}
	sort.Slice(queued, func(i, j int) bool { return queued[i].id < queued[j].id })
	a.mu.Unlock()
	defer a.endWork()
	for _, op := range queued {
		a.mu.Lock()
		if a.stopping {
			a.mu.Unlock()
			break
		}
		if op.released || !op.desired || op.phase != phaseQueued {
			a.mu.Unlock()
			continue
		}
		attempt := a.claimStartLocked(op, context.Background(), "registry")
		a.mu.Unlock()
		_ = a.joinStart(op, attempt)
	}
	return a.SaveRegistry()
}

func (a *Agent) SaveRegistry() error {
	a.registryWrite.Lock()
	defer a.registryWrite.Unlock()
	if err := a.imported(); err != nil {
		return fmt.Errorf("registry was not imported; preserving unread file: %w", err)
	}
	a.mu.Lock()
	registry := &Registry{V: RegistryV}
	var ids []string
	for id, op := range a.operations {
		if op.desired && !op.released {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	for _, id := range ids {
		op := a.operations[id]
		var failure *FailureEntry
		if op.failure != nil {
			copy := *op.failure
			failure = &copy
		}
		switch req := cloneRequest(op.request).(type) {
		case *SharePublicRequest:
			registry.PublicShares = append(registry.PublicShares, &PublicShareRegistryEntry{Request: req, Failure: failure})
		case *SharePrivateRequest:
			registry.PrivateShares = append(registry.PrivateShares, &PrivateShareRegistryEntry{Request: req, Failure: failure})
		case *AccessPrivateRequest:
			registry.PrivateAccesses = append(registry.PrivateAccesses, &AccessRegistryEntry{Request: req, Failure: failure})
		}
	}
	a.mu.Unlock()
	path, err := a.root.AgentRegistry()
	if err != nil {
		return err
	}
	return a.deps.saveRegistry(registry, path)
}
