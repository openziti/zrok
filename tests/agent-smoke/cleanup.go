//go:build agent_smoke

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/michaelquigley/df/dl"
)

func cleanupMode(ctx context.Context, cfg config) int {
	rep := newReport(cfg.stateDir, "cleanup", cfg.accountToken)
	st, err := loadRunState(cfg.stateDir)
	if errors.Is(err, os.ErrNotExist) {
		rep.cleanup("no manifest; the run recorded no resources", nil, 0)
		return finish(rep, manifest{})
	}
	if err != nil {
		rep.cleanup("", err, 0)
		return finish(rep, manifest{})
	}
	h := &harness{cfg: cfg, st: st, rep: rep}
	h.runCleanup(ctx)
	return finish(rep, st.snapshot())
}

// runCleanup gives cleanup its own allowance, independent of whatever ended the scenario.
func (h *harness) runCleanup(parent context.Context) {
	ctx, cancel := context.WithTimeout(parent, h.cfg.cleanupBound)
	defer cancel()
	started := time.Now()
	detail, err := h.cleanup(ctx)
	h.rep.cleanup(detail, err, time.Since(started))
}

// cleanup stops this run's processes and removes exactly the resources the manifest records, then proves their
// absence with successful authenticated reads. it is safe to repeat. intermediate write failures are notes; the
// verification decides. an unresolved process fails cleanup without stopping the controller cleanup.
func (h *harness) cleanup(ctx context.Context) (_ string, err error) {
	var notes []string
	note := func(format string, args ...any) {
		text := fmt.Sprintf(format, args...)
		dl.Infof("cleanup: %v", text)
		notes = append(notes, text)
	}
	detail := func() string { return strings.Join(notes, "; ") }

	if h.target != nil {
		h.target.close()
	}
	if h.agentConn != nil {
		_ = h.agentConn.Close()
	}
	processErr := h.stopProcesses(ctx, note)
	defer func() { err = errors.Join(processErr, err) }()

	m := h.st.snapshot()
	if !m.EnvAttempted && !m.NameAttempted && len(m.ShareTokens) == 0 {
		note("no controller resources recorded")
		return detail(), nil
	}
	if h.cfg.accountToken == "" {
		return detail(), fmt.Errorf("'%v' is not set; recorded resources remain: %v", accountTokenEnv, recordedResources(m))
	}
	if h.ctrl == nil {
		ctrl, err := newControllerClient(ctx, m.APIEndpoint, h.cfg.accountToken, h.cfg.requestBound)
		if err != nil {
			return detail(), fmt.Errorf("%w; recorded resources remain: %v", err, recordedResources(m))
		}
		h.ctrl = ctrl
	}
	view, err := h.readView(ctx)
	if err != nil {
		return detail(), fmt.Errorf("%w; recorded resources remain: %v", err, recordedResources(m))
	}

	// resolve the environment: the recorded identity, or for a lost enable response the unique description
	var problems []error
	zID := m.EnvZID
	if zID == "" && m.EnvAttempted {
		described := view.environmentsDescribed(m.Description)
		switch len(described) {
		case 0:
		case 1:
			zID = described[0]
			if err := h.st.update(func(m *manifest) { m.EnvZID = zID }); err != nil {
				return detail(), err
			}
			note("reconciled environment '%v' by description", zID)
		default:
			problems = append(problems, fmt.Errorf("environments '%v' all carry description '%v'; none removed", strings.Join(described, ", "), m.Description))
		}
	}
	owned := false
	if zID != "" {
		if env := view.environment(zID); env != nil {
			if env.Environment.Description != m.Description {
				problems = append(problems, fmt.Errorf("environment '%v' has description '%v', not this run's '%v'; not removed", zID, env.Environment.Description, m.Description))
			} else {
				owned = true
				for _, shr := range env.Shares {
					if shr == nil || shr.ShareToken == "" {
						continue
					}
					if err := h.st.addShareToken(shr.ShareToken); err != nil {
						return detail(), err
					}
					if err := h.ctrl.unshare(ctx, zID, shr.ShareToken); err != nil {
						note("%v", err)
					} else {
						note("unshared '%v'", shr.ShareToken)
					}
				}
			}
		}
	}

	// the name was absent before this run reserved it, so a listing of the exact name is this run's
	nameOwned := m.NameAttempted && !m.NameForeign
	retryName := false
	if nameOwned && len(view.matchingNames(m.Namespace, m.Name)) > 0 {
		if err := h.ctrl.deleteName(ctx, m.Namespace, m.Name); err != nil {
			note("%v; retrying after disable", err)
			retryName = true
		} else {
			note("deleted name '%v:%v'", m.Namespace, m.Name)
		}
	}
	if owned {
		if err := h.ctrl.disable(ctx, zID); err != nil {
			note("%v", err)
		} else {
			note("disabled environment '%v'", zID)
		}
	}
	if retryName {
		if err := h.ctrl.deleteName(ctx, m.Namespace, m.Name); err != nil {
			note("%v", err)
		} else {
			note("deleted name '%v:%v'", m.Namespace, m.Name)
		}
	}

	// verify with a fresh successful read
	m = h.st.snapshot()
	after, err := h.readView(ctx)
	if err != nil {
		return detail(), errors.Join(append(problems, fmt.Errorf("%w; absence unverified for: %v", err, recordedResources(m)))...)
	}
	if zID != "" && after.environment(zID) != nil {
		problems = append(problems, fmt.Errorf("environment '%v' is still present", zID))
	}
	if m.EnvAttempted {
		for _, other := range after.environmentsDescribed(m.Description) {
			if other != zID {
				problems = append(problems, fmt.Errorf("environment '%v' with this run's description is still present", other))
			}
		}
	}
	for _, token := range m.ShareTokens {
		if locations := after.tokenLocations(token); len(locations) > 0 {
			problems = append(problems, fmt.Errorf("share '%v' is still present: %v", token, strings.Join(locations, "; ")))
		}
	}
	if nameOwned && len(after.matchingNames(m.Namespace, m.Name)) > 0 {
		problems = append(problems, fmt.Errorf("name '%v:%v' is still listed", m.Namespace, m.Name))
	}
	if len(problems) == 0 {
		note("verified absence of %v", recordedResources(m))
	}
	return detail(), errors.Join(problems...)
}

// stopProcesses force-terminates what an agent this invocation owns still has running. a process recorded by
// another invocation is acted on only while it still shows the recorded identity, and only by a build for the
// platform that recorded it. a record is marked reaped only once its process is confirmed gone; every other
// outcome is returned and the record is kept for a later cleanup.
func (h *harness) stopProcesses(ctx context.Context, note func(string, ...any)) error {
	var problems []error
	if h.agent != nil {
		if killed, err := h.agent.kill(h.cfg.killBound); err != nil {
			problems = append(problems, err)
		} else if killed != "" {
			note("force-terminated agent %v", killed)
		}
		if h.agent.exited() {
			if err := h.markReaped(h.agent.pid()); err != nil {
				problems = append(problems, err)
			}
		}
	}
	for _, record := range h.st.snapshot().Processes {
		if record.Reaped || (h.agent != nil && record.PID == h.agent.pid()) {
			continue
		}
		if platform := recordPlatform(record); platform != localPlatform {
			problems = append(problems, fmt.Errorf("recorded agent pid %d has a '%v' identity and cannot be inspected on '%v'; record kept", record.PID, platform, localPlatform))
			continue
		}
		outcome, err := stopRecorded(ctx, record, h.cfg.killBound, h.cfg.pollInterval)
		var inspect *inspectError
		switch {
		case errors.As(err, &inspect):
			problems = append(problems, fmt.Errorf("unable to inspect recorded agent pid %d: %w; record kept", record.PID, inspect.err))
			continue
		case err != nil:
			problems = append(problems, fmt.Errorf("recorded agent pid %d %w; record kept", record.PID, err))
			continue
		case outcome == recordNotRunning:
			note("recorded agent pid %d is not running", record.PID)
		case outcome == recordReplaced:
			note("recorded pid %d is now a different process; not signalled", record.PID)
		default:
			note("force-terminated recorded agent pid %d", record.PID)
		}
		if err := h.markReaped(record.PID); err != nil {
			problems = append(problems, err)
		}
	}
	return errors.Join(problems...)
}

func recordedResources(m manifest) string {
	parts := []string{fmt.Sprintf("description '%v'", m.Description)}
	if m.EnvZID != "" {
		parts = append(parts, fmt.Sprintf("environment '%v'", m.EnvZID))
	}
	if m.NameAttempted && !m.NameForeign {
		parts = append(parts, fmt.Sprintf("name '%v:%v'", m.Namespace, m.Name))
	}
	for _, token := range m.ShareTokens {
		parts = append(parts, fmt.Sprintf("share '%v'", token))
	}
	return strings.Join(parts, ", ")
}
