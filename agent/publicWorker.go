package agent

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/agent/proctree"
	"github.com/openziti/zrok/v2/cmd/zrok2/subordinate"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
)

// publicCreateError distinguishes a refused create from a later backend or cleanup failure.
type publicCreateError struct{ cause error }

func (e *publicCreateError) Error() string { return e.cause.Error() }
func (e *publicCreateError) Unwrap() error { return e.cause }

func (a *Agent) launchPublic(op *operation, attempt *startAttempt, req *SharePublicRequest) (*share, error) {
	prepared, err := a.preparePublicRequest(req)
	if err != nil {
		return nil, err
	}
	if err := errors.Join(attempt.ctx.Err(), attempt.caller.Err()); err != nil {
		return nil, err
	}
	// once dispatched, let the bounded request settle even if the caller or agent stops.
	// a returned token must enter the registered operation before any cancellation is handled.
	ctx, cancel := context.WithTimeout(context.Background(), a.options.create)
	created, err := a.deps.createPublic(ctx, prepared)
	cancel()
	if err != nil {
		return nil, &publicCreateError{cause: err}
	}
	if created == nil || strings.TrimSpace(created.Token) == "" {
		return nil, errors.New("controller returned a public share without a token")
	}
	var names []NameSelection
	for _, selection := range prepared.NameSelections {
		names = append(names, NameSelection{NamespaceToken: selection.NamespaceToken, Name: selection.Name})
	}
	shr := &share{token: created.Token, frontendEndpoints: slices.Clone(created.FrontendEndpoints),
		shareMode: sdk.PublicShareMode, backendMode: prepared.BackendMode, target: prepared.Target,
		basicAuth: prepared.BasicAuth, nameSelections: names, insecure: req.Insecure,
		oauthProvider: prepared.OauthProvider, oauthEmailAddressPatterns: prepared.OauthEmailAddressPatterns,
		closed: req.Closed, accessGrants: prepared.AccessGrants}
	a.mu.Lock()
	op.shr = shr
	stopping := op.stop || op.released || a.stopping
	a.mu.Unlock()
	if err := errors.Join(attempt.ctx.Err(), attempt.caller.Err()); err != nil {
		return shr, err
	}
	if stopping {
		return shr, errors.New("public share stopped during creation")
	}
	boot := &legacyBoot{shr: shr}
	handler := NewPublicWorkerBootHandler(shr, &boot.err)
	sub := subordinate.NewMessageHandler()
	sub.BootHandler = func(kind string, msg subordinate.Message) {
		boot.mu.Lock()
		defer boot.mu.Unlock()
		handler.HandleBoot(kind, msg)
	}
	sub.MalformedHandler = handler.HandleMalformed
	sub.MessageHandler = func(msg subordinate.Message) { dl.Info(msg) }
	command := NewSharePublicCommand().ShareToken(shr.token).BackendMode(string(prepared.BackendMode)).Insecure(req.Insecure).Target(prepared.Target).Build()
	startup, stop := context.WithTimeout(attempt.ctx, a.options.startup)
	defer stop()
	child, err := a.deps.startPublic(proctree.ControlledOptions{
		Stdout: sub.Tail, StdoutClosed: sub.Close,
		Stderr: func(data []byte) { dl.Infof("public worker stderr: '%s'", data) },
	}, command...)
	if err != nil {
		return shr, err
	}
	a.mu.Lock()
	op.child = child
	a.mu.Unlock()
	select {
	case <-sub.BootComplete:
		_, _, err := boot.snapshot()
		select {
		case <-child.Done():
			err = errors.Join(err, errors.New("public worker exited during startup"), proctree.WaitChild(child))
		default:
		}
		return shr, err
	case <-child.Done():
		return shr, fmt.Errorf("public worker exited before startup completed: %w", errors.Join(proctree.WaitChild(child), errors.New("worker is not running")))
	case <-startup.Done():
		return shr, startup.Err()
	}
}
