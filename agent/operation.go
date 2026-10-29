package agent

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"sync"
	"time"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/agent/proctree"
	"github.com/openziti/zrok/v2/cmd/zrok2/subordinate"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
)

type operationPhase string

const (
	phaseQueued   operationPhase = "queued"
	phaseStarting operationPhase = "starting"
	phaseActive   operationPhase = "active"
	phaseStopping operationPhase = "stopping"
	phaseRetry    operationPhase = "retrying"
	phaseFailed   operationPhase = "failed"
	phaseIdle     operationPhase = "idle"
)

type startAttempt struct {
	ctx    context.Context
	cancel context.CancelFunc
	caller context.Context
	detach func() bool
	done   chan struct{}
	err    error
	shr    *share
	acc    *access
}

type cleanupAttempt struct {
	done chan struct{}
	err  error
}

// operation is owned by Agent.mu; a request remains identifiable across startup, cleanup and retry.
type operation struct {
	id             string
	origin         string
	request        any
	phase          operationPhase
	desired        bool
	released       bool
	stop           bool
	retryAfterStop bool
	failureID      string
	failure        *FailureEntry
	lastError      error
	start          *startAttempt
	cleanup        *cleanupAttempt
	child          *proctree.Child
	shr            *share
	acc            *access
	bootSnapshot   func() (*share, *access, error)
}

type lifecycleOptions struct {
	create     time.Duration
	startup    time.Duration
	graceful   time.Duration
	forcedExit time.Duration
	cleanup    time.Duration
	shutdown   time.Duration
	server     time.Duration
}

func defaultLifecycleOptions() lifecycleOptions {
	return lifecycleOptions{create: time.Minute, startup: 6 * time.Minute, graceful: 5 * time.Second, forcedExit: 5 * time.Second, cleanup: 30 * time.Second, shutdown: 120 * time.Second, server: 5 * time.Second}
}

func cloneRequest(request any) any {
	switch in := request.(type) {
	case *SharePublicRequest:
		out := *in
		out.BasicAuth = slices.Clone(in.BasicAuth)
		out.NameSelections = slices.Clone(in.NameSelections)
		out.OauthEmailDomains = slices.Clone(in.OauthEmailDomains)
		out.AccessGrants = slices.Clone(in.AccessGrants)
		return &out
	case *SharePrivateRequest:
		out := *in
		out.AccessGrants = slices.Clone(in.AccessGrants)
		return &out
	case *AccessPrivateRequest:
		out := *in
		out.ResponseHeaders = slices.Clone(in.ResponseHeaders)
		return &out
	default:
		panic("unsupported agent request")
	}
}

// sameRequest treats nil and empty transport lists alike without changing saved or caller-owned data.
func sameRequest(left, right any) bool {
	switch in := left.(type) {
	case *SharePublicRequest:
		other, ok := right.(*SharePublicRequest)
		if !ok {
			return false
		}
		a, b := *in, *other
		a.BasicAuth, b.BasicAuth = nil, nil
		a.NameSelections, b.NameSelections = nil, nil
		a.OauthEmailDomains, b.OauthEmailDomains = nil, nil
		a.AccessGrants, b.AccessGrants = nil, nil
		return reflect.DeepEqual(a, b) && slices.Equal(in.BasicAuth, other.BasicAuth) && slices.Equal(in.NameSelections, other.NameSelections) && slices.Equal(in.OauthEmailDomains, other.OauthEmailDomains) && slices.Equal(in.AccessGrants, other.AccessGrants)
	case *SharePrivateRequest:
		other, ok := right.(*SharePrivateRequest)
		if !ok {
			return false
		}
		a, b := *in, *other
		a.AccessGrants, b.AccessGrants = nil, nil
		return reflect.DeepEqual(a, b) && slices.Equal(in.AccessGrants, other.AccessGrants)
	case *AccessPrivateRequest:
		other, ok := right.(*AccessPrivateRequest)
		if !ok {
			return false
		}
		a, b := *in, *other
		a.ResponseHeaders, b.ResponseHeaders = nil, nil
		return reflect.DeepEqual(a, b) && slices.Equal(in.ResponseHeaders, other.ResponseHeaders)
	default:
		return false
	}
}

func restartEligible(request any) bool {
	switch req := request.(type) {
	case *SharePublicRequest:
		return req.hasReservedName()
	case *SharePrivateRequest:
		return req.hasReservedToken()
	case *AccessPrivateRequest:
		return true
	}
	return false
}

func operationID() string { return "err_" + rand.Text() }

func (a *Agent) beginWorkLocked() {
	if a.workers == 0 {
		a.idle = make(chan struct{})
	}
	a.workers++
}
func (a *Agent) endWork() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.workers--
	if a.workers == 0 {
		close(a.idle)
	}
}

func (a *Agent) claimStartLocked(op *operation, caller context.Context, origin string) *startAttempt {
	ctx, cancel := context.WithTimeout(a.ctx, a.options.startup)
	if _, public := op.request.(*SharePublicRequest); public {
		cancel()
		ctx, cancel = context.WithCancel(a.ctx)
	}
	attempt := &startAttempt{ctx: ctx, cancel: cancel, caller: caller, done: make(chan struct{})}
	attempt.detach = context.AfterFunc(caller, cancel)
	op.start = attempt
	op.cleanup = nil
	op.phase = phaseStarting
	op.stop = false
	op.lastError = nil
	op.origin = origin
	a.beginWorkLocked()
	go a.runStart(op, attempt)
	return attempt
}

func (a *Agent) startRequest(ctx context.Context, request any) (*startAttempt, error) {
	if err := a.waitForImport(ctx); err != nil {
		return nil, err
	}
	a.mu.Lock()
	if a.stopping {
		a.mu.Unlock()
		return nil, errors.New("agent is stopping")
	}
	if err := ctx.Err(); err != nil {
		a.mu.Unlock()
		return nil, err
	}
	var op *operation
	for _, queued := range a.operations {
		if queued.phase == phaseQueued && !queued.released && sameRequest(queued.request, request) {
			op = queued
			break
		}
	}
	if op == nil {
		op = &operation{id: operationID(), request: cloneRequest(request)}
		a.operations[op.id] = op
	}
	attempt := a.claimStartLocked(op, ctx, "caller")
	a.mu.Unlock()
	return attempt, a.joinStart(op, attempt)
}

func (a *Agent) joinStart(op *operation, attempt *startAttempt) error {
	<-attempt.done
	a.mu.Lock()
	err := attempt.err
	cleanup := op.cleanup
	a.mu.Unlock()
	if err != nil && cleanup != nil {
		<-cleanup.done
		err = errors.Join(err, cleanup.err)
	}
	return err
}

func (a *Agent) runStart(op *operation, attempt *startAttempt) {
	defer a.endWork()
	var shr *share
	var acc *access
	var err error
	if req, public := op.request.(*SharePublicRequest); public {
		shr, err = a.launchPublic(op, attempt, req)
	} else {
		shr, acc, err = a.launchLegacy(op, attempt)
	}
	attempt.detach()
	a.mu.Lock()
	op.shr = shr
	op.acc = acc
	if err == nil {
		err = attempt.ctx.Err()
	}
	if err == nil {
		err = attempt.caller.Err()
	}
	if err == nil && (op.stop || op.released || a.stopping) {
		err = errors.New("operation stopped during startup")
	}
	attempt.err = err
	attempt.shr = shr
	attempt.acc = acc
	if err == nil {
		op.phase = phaseActive
		op.failure = nil
		op.failureID = ""
		op.desired = restartEligible(op.request)
		a.beginWorkLocked()
		child := op.child
		go a.monitorOperation(op, attempt, child)
	} else {
		op.stop = true
		op.retryAfterStop = op.desired && !op.released
	}
	close(attempt.done)
	if err != nil {
		a.claimCleanupLocked(op)
	}
	a.mu.Unlock()
	attempt.cancel()
	if err := a.SaveRegistry(); err != nil {
		dl.Errorf("unable to save agent registry: %v", err)
	}
}

func (a *Agent) monitorOperation(op *operation, attempt *startAttempt, child *proctree.Child) {
	defer a.endWork()
	err := proctree.WaitChild(child)
	a.mu.Lock()
	if op.start == attempt && op.phase == phaseActive {
		if err == nil {
			err = errors.New("worker exited unexpectedly")
		}
		op.lastError = err
		op.retryAfterStop = op.desired && !op.released
		op.stop = true
		a.claimCleanupLocked(op)
	}
	a.mu.Unlock()
}

func (a *Agent) claimCleanupLocked(op *operation) *cleanupAttempt {
	if op.cleanup != nil {
		select {
		case <-op.cleanup.done:
			if op.cleanup.err == nil {
				if op.released && op.child == nil && op.resourceToken() == "" {
					delete(a.operations, op.id)
					op.phase = phaseIdle
				}
				return op.cleanup
			}
		default:
			return op.cleanup
		}
	}
	attempt := &cleanupAttempt{done: make(chan struct{})}
	op.cleanup = attempt
	op.phase = phaseStopping
	a.beginWorkLocked()
	go a.runCleanup(op, attempt)
	return attempt
}

func (a *Agent) runCleanup(op *operation, attempt *cleanupAttempt) {
	defer a.endWork()
	a.mu.Lock()
	start := op.start
	a.mu.Unlock()
	if start != nil {
		<-start.done
	}
	a.mu.Lock()
	child := op.child
	snapshot := op.bootSnapshot
	_, public := op.request.(*SharePublicRequest)
	a.mu.Unlock()
	var err error
	if child != nil {
		if public {
			err = proctree.StopControlledChild(context.Background(), child, a.options.graceful, a.options.forcedExit)
		} else {
			err = proctree.StopChild(child)
		}
		if err == nil && !public {
			ctx, cancel := context.WithTimeout(context.Background(), a.options.forcedExit)
			select {
			case <-child.Done():
			case <-ctx.Done():
				err = fmt.Errorf("worker exit is unconfirmed: %w", ctx.Err())
			}
			cancel()
		}
		if err == nil {
			<-child.DrainWaitDone()
		}
	}
	if err == nil && snapshot != nil {
		shr, acc, _ := snapshot()
		a.mu.Lock()
		op.shr = shr
		op.acc = acc
		a.mu.Unlock()
	}
	a.mu.Lock()
	shr, acc := op.shr, op.acc
	released := op.released
	a.mu.Unlock()
	shareDeleted := false
	if err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), a.options.cleanup)
		if acc != nil && acc.frontendToken != "" {
			err = a.deps.deleteAccess(ctx, acc.token, acc.frontendToken)
		}
		if shr != nil && shr.token != "" && (public || released || !restartEligible(op.request)) {
			shareDeleted = true
			if public {
				err = a.deps.deletePublic(ctx, shr.token)
			} else {
				err = a.deps.deleteShare(ctx, shr.token)
			}
		}
		cancel()
	}
	a.mu.Lock()
	// a release arriving during the stop joins this attempt and upgrades a reserved share's deletion policy.
	if err == nil && op.released && !shareDeleted && shr != nil && shr.token != "" {
		a.mu.Unlock()
		ctx, cancel := context.WithTimeout(context.Background(), a.options.cleanup)
		err = a.deps.deleteShare(ctx, shr.token)
		cancel()
		a.mu.Lock()
	}
	if err != nil {
		err = fmt.Errorf("pending cleanup for '%v': %w", op.resourceToken(), err)
		a.recordFailureLocked(op, err)
		op.phase = phaseFailed
	} else {
		op.child = nil
		op.shr = nil
		op.acc = nil
		op.bootSnapshot = nil
		if op.retryAfterStop && op.desired && !op.released && !a.stopping {
			cause := errors.New("worker stopped")
			if op.lastError != nil {
				cause = op.lastError
			} else if start != nil && start.err != nil {
				cause = start.err
			}
			a.recordFailureLocked(op, cause)
			op.phase = phaseRetry
		} else {
			op.phase = phaseIdle
		}
		if !op.desired {
			delete(a.operations, op.id)
		}
	}
	a.mu.Unlock()
	saveErr := a.SaveRegistry()
	a.mu.Lock()
	// release may have joined this attempt while persistence ran outside the lifecycle mutex.
	if op.released && op.child == nil && op.resourceToken() == "" {
		delete(a.operations, op.id)
		op.phase = phaseIdle
	}
	attempt.err = errors.Join(err, saveErr)
	close(attempt.done)
	a.mu.Unlock()
}

func (op *operation) resourceToken() string {
	if op.shr != nil {
		return op.shr.token
	}
	if op.acc != nil {
		return op.acc.frontendToken
	}
	return ""
}

func (a *Agent) recordFailureLocked(op *operation, err error) {
	if op.failureID == "" {
		op.failureID = op.id
	}
	count := 1
	if op.failure != nil {
		count = op.failure.Count + 1
	}
	op.failure = &FailureEntry{Count: count, LastError: err.Error(), NextRetry: a.retryCalc.nextRetryTime(count)}
}

// legacyBoot isolates parser writes from lifecycle snapshots, including a boot that arrives after cancellation.
type legacyBoot struct {
	mu  sync.Mutex
	shr *share
	acc *access
	err error
}

func (b *legacyBoot) snapshot() (*share, *access, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	var shr *share
	var acc *access
	if b.shr != nil {
		out := *b.shr
		out.frontendEndpoints = slices.Clone(out.frontendEndpoints)
		out.basicAuth = slices.Clone(out.basicAuth)
		out.nameSelections = slices.Clone(out.nameSelections)
		out.accessGrants = slices.Clone(out.accessGrants)
		out.oauthEmailAddressPatterns = slices.Clone(out.oauthEmailAddressPatterns)
		shr = &out
	}
	if b.acc != nil {
		out := *b.acc
		out.responseHeaders = slices.Clone(out.responseHeaders)
		acc = &out
	}
	return shr, acc, b.err
}

func (a *Agent) launchLegacy(op *operation, attempt *startAttempt) (*share, *access, error) {
	boot := &legacyBoot{}
	var handler *UnifiedBootHandler
	var command []string
	switch req := op.request.(type) {
	case *SharePrivateRequest:
		boot.shr = &share{shareMode: sdk.PrivateShareMode, backendMode: sdk.BackendMode(req.BackendMode), target: req.Target, insecure: req.Insecure, closed: req.Closed, accessGrants: req.AccessGrants}
		builder := NewSharePrivateCommand()
		if req.PrivateShareToken == "" {
			builder = builder.BackendMode(req.BackendMode)
		}
		command = builder.ShareToken(req.PrivateShareToken).Insecure(req.Insecure).Open(!req.Closed).AccessGrants(req.AccessGrants).Target(req.Target).Build()
		handler = NewShareBootHandler(boot.shr, &boot.err)
	case *AccessPrivateRequest:
		boot.acc = &access{token: req.ShareToken, bindAddress: req.BindAddress, autoMode: req.AutoMode, autoAddress: req.AutoAddress, autoStartPort: req.AutoStartPort, autoEndPort: req.AutoEndPort, responseHeaders: req.ResponseHeaders}
		command = NewAccessPrivateCommand().BindAddress(req.BindAddress).AutoMode(req.AutoMode, req.AutoAddress, int(req.AutoStartPort), int(req.AutoEndPort)).Target(req.ShareToken).Build()
		handler = NewAccessBootHandler(boot.acc, &boot.err)
	}
	sub := subordinate.NewMessageHandler()
	sub.BootHandler = func(kind string, msg subordinate.Message) {
		boot.mu.Lock()
		defer boot.mu.Unlock()
		handler.HandleBoot(kind, msg)
	}
	sub.MalformedHandler = handler.HandleMalformed
	sub.MessageHandler = func(msg subordinate.Message) { dl.Info(msg) }
	if err := attempt.ctx.Err(); err != nil {
		return nil, nil, err
	}
	child, err := a.deps.startChild(sub.Tail, command...)
	if err != nil {
		return nil, nil, err
	}
	a.mu.Lock()
	op.child = child
	op.bootSnapshot = boot.snapshot
	a.mu.Unlock()
	select {
	case <-sub.BootComplete:
		shr, acc, err := boot.snapshot()
		return shr, acc, err
	case <-child.Done():
		shr, acc, _ := boot.snapshot()
		return shr, acc, fmt.Errorf("worker exited before startup completed: %w", errors.Join(proctree.WaitChild(child), errors.New("worker is not running")))
	case <-attempt.ctx.Done():
		shr, acc, _ := boot.snapshot()
		return shr, acc, attempt.ctx.Err()
	}
}
