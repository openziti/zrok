//go:build agent_smoke && unix

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/agent/agentGrpc"
	"github.com/openziti/zrok/v2/build"
	"github.com/openziti/zrok/v2/environment"
	"github.com/openziti/zrok/v2/environment/env_core"
	"google.golang.org/grpc"
)

// harness holds one invocation's state. run mode owns the processes it starts; cleanup mode owns none and acts
// only on the manifest.
type harness struct {
	cfg  config
	st   *runState
	rep  *report
	ctrl *controllerClient

	root      env_core.Root
	target    *loopback
	agent     *child
	agentRPC  agentGrpc.AgentClient
	agentConn *grpc.ClientConn
	starts    int
	tokens    []string
}

func runMode(first, second context.Context, cfg config) int {
	rep := newReport(cfg.stateDir, "run", cfg.accountToken)
	unstarted := manifest{RunID: cfg.runID, RunAttempt: cfg.runAttempt, APIEndpoint: cfg.apiEndpoint, Namespace: cfg.namespace}
	if err := preflight(cfg); err != nil {
		rep.fail(fmt.Errorf("preflight: %w", err))
		return finish(rep, unstarted)
	}
	st, err := newRunState(cfg.stateDir, newManifest(cfg))
	if err != nil {
		rep.fail(fmt.Errorf("preflight: %w", err))
		return finish(rep, unstarted)
	}
	m := st.snapshot()
	dl.Infof("manifest '%v' records name '%v:%v' and environment description '%v' at '%v'", st.path, m.Namespace, m.Name, m.Description, m.APIEndpoint)

	h := &harness{cfg: cfg, st: st, rep: rep}
	scenarioCtx, cancel := context.WithTimeout(first, cfg.scenarioBound)
	if err := h.scenario(scenarioCtx); err != nil {
		rep.fail(err)
	}
	cancel()

	h.runCleanup(second)
	return finish(rep, st.snapshot())
}

func finish(rep *report, m manifest) int {
	if err := rep.write(m); err != nil {
		dl.Errorf("unable to write report: %v", err)
		return 1
	}
	if rep.failed() {
		return 1
	}
	return 0
}

func preflight(cfg config) error {
	if cfg.accountToken == "" {
		return fmt.Errorf("'%v' is not set; the live smoke requires the dedicated account token", accountTokenEnv)
	}
	if build.Version == "" || build.Hash == "" {
		return errors.New("the harness must be built with the same build.Version and build.Hash ldflags as the binary under test")
	}
	if cfg.namespace == "" {
		return errors.New("a namespace is required")
	}
	info, err := os.Stat(cfg.binary)
	if err != nil {
		return fmt.Errorf("binary under test: %w", err)
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0o111 == 0 {
		return fmt.Errorf("binary under test '%v' is not an executable file", cfg.binary)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	if _, err := os.Lstat(filepath.Join(home, ".zrok2")); err == nil {
		return fmt.Errorf("'%v' already exists; the smoke runs only in a fresh home and never adopts an existing environment", filepath.Join(home, ".zrok2"))
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

var unsafeIdentity = regexp.MustCompile(`[^a-z0-9]+`)

// identitySuffix is the random part of a run's identity; tests replace it to arrange a name conflict.
var identitySuffix = func() string { return randomHex(4) }

// newManifest chooses this run's unique name and environment description before any side effect.
func newManifest(cfg config) manifest {
	run := unsafeIdentity.ReplaceAllString(strings.ToLower(cfg.runID+"-"+cfg.runAttempt), "-")
	suffix := identitySuffix()
	const prefix = "agentsmoke-"
	if limit := 63 - len(prefix) - 1 - len(suffix); len(run) > limit {
		run = run[len(run)-limit:]
	}
	identity := strings.Trim(run, "-") + "-" + suffix
	return manifest{
		RunID:       cfg.runID,
		RunAttempt:  cfg.runAttempt,
		APIEndpoint: cfg.apiEndpoint,
		Namespace:   cfg.namespace,
		Name:        prefix + identity,
		Description: "zrok agent smoke " + identity,
	}
}

func (h *harness) scenario(ctx context.Context) error {
	steps := []struct {
		name string
		fn   func(context.Context) (string, error)
	}{
		{"controller version check", h.connectController},
		{"enable environment", h.enable},
		{"reserve name", h.reserveName},
		{"start loopback target", h.startTarget},
		{"start agent", h.startAgent},
		{"share through agent", h.share},
		{"public response", h.checkPublic},
		{"interrupt agent", h.stopAgent},
		{"controller state after stop", h.checkStopped},
		{"restart agent", h.startAgent},
		{"replayed share", h.awaitReplay},
		{"public response after replay", h.checkPublic},
		{"interrupt restarted agent", h.stopAgent},
		{"controller state after second stop", h.checkStopped},
	}
	for _, step := range steps {
		if err := h.rep.step(step.name, func() (string, error) { return step.fn(ctx) }); err != nil {
			return err
		}
	}
	return nil
}

func (h *harness) connectController(ctx context.Context) (string, error) {
	ctrl, err := newControllerClient(ctx, h.st.snapshot().APIEndpoint, h.cfg.accountToken, h.cfg.requestBound)
	if err != nil {
		return "", err
	}
	h.ctrl = ctrl
	return fmt.Sprintf("accepted '%v'", build.String()), nil
}

func (h *harness) enable(ctx context.Context) (string, error) {
	m := h.st.snapshot()
	if err := h.st.update(func(m *manifest) { m.EnvAttempted = true }); err != nil {
		return "", err
	}
	if err := h.runCommand(ctx, "enable", h.cfg.commandBound, "enable", "--headless", "-d", m.Description, h.cfg.accountToken); err != nil {
		return "", err
	}
	root, err := environment.LoadRoot()
	if err != nil {
		return "", fmt.Errorf("unable to load the enabled environment: %w", err)
	}
	if !root.IsEnabled() || root.Environment().ZitiIdentity == "" {
		return "", errors.New("enable succeeded but no enabled environment was written")
	}
	zID := root.Environment().ZitiIdentity
	if err := h.st.update(func(m *manifest) { m.EnvZID = zID }); err != nil {
		return "", err
	}
	dl.Infof("recorded environment '%v'", zID)
	h.root = root

	view, err := h.readView(ctx)
	if err != nil {
		return "", err
	}
	env := view.environment(zID)
	if env == nil {
		return "", fmt.Errorf("environment '%v' is missing from the account overview", zID)
	}
	if env.Environment.Description != m.Description {
		return "", fmt.Errorf("environment '%v' has description '%v'; expected '%v'", zID, env.Environment.Description, m.Description)
	}
	return fmt.Sprintf("environment '%v'", zID), nil
}

func (h *harness) reserveName(ctx context.Context) (string, error) {
	m := h.st.snapshot()
	names, err := h.ctrl.names(ctx)
	if err != nil {
		return "", err
	}
	if (&accountView{names: names}).matchingNames(m.Namespace, m.Name) != nil {
		if err := h.st.update(func(m *manifest) { m.NameForeign = true }); err != nil {
			return "", err
		}
		return "", fmt.Errorf("name '%v:%v' is already listed for this account; it is not adopted or deleted", m.Namespace, m.Name)
	}
	if err := h.st.update(func(m *manifest) { m.NameAttempted = true }); err != nil {
		return "", err
	}
	if err := h.runCommand(ctx, "create-name", h.cfg.commandBound, "create", "name", "-n", m.Namespace, m.Name); err != nil {
		return "", err
	}
	if err := h.st.update(func(m *manifest) { m.NameReserved = true }); err != nil {
		return "", err
	}
	dl.Infof("recorded name '%v:%v'", m.Namespace, m.Name)
	view, err := h.readView(ctx)
	if err != nil {
		return "", err
	}
	if err := requireName(view, m, ""); err != nil {
		return "", err
	}
	return fmt.Sprintf("'%v:%v' reserved and unbound", m.Namespace, m.Name), nil
}

func (h *harness) startTarget(_ context.Context) (string, error) {
	lb, err := startLoopback()
	if err != nil {
		return "", err
	}
	h.target = lb
	return fmt.Sprintf("serving at '%v'", lb.target()), nil
}

func (h *harness) startAgent(ctx context.Context) (string, error) {
	h.starts++
	label := fmt.Sprintf("agent-%d", h.starts)
	agent, err := startChild(h.cfg.binary, []string{"agent", "start"}, h.childEnv(), h.rep.log(label))
	if err != nil {
		return "", fmt.Errorf("unable to start agent: %w", err)
	}
	h.agent = agent
	record := processRecord{PID: agent.pid(), Command: h.cfg.binary + " agent start"}
	if started, _, alive, err := processIdentity(ctx, agent.pid()); err == nil && alive {
		record.Started = started
	}
	if err := h.st.update(func(m *manifest) { m.Processes = append(m.Processes, record) }); err != nil {
		return "", err
	}
	dl.Infof("recorded agent pid '%d'", agent.pid())

	startCtx, cancel := context.WithTimeout(ctx, h.cfg.startupBound)
	defer cancel()
	client, conn, err := awaitVersion(startCtx, h.cfg, h.root, agent, build.String())
	if err != nil {
		return "", err
	}
	h.agentRPC, h.agentConn = client, conn
	return fmt.Sprintf("pid %d reported '%v'", agent.pid(), build.String()), nil
}

func (h *harness) share(ctx context.Context) (string, error) {
	m := h.st.snapshot()
	shareCtx, cancel := context.WithTimeout(ctx, h.cfg.startupBound)
	defer cancel()
	selection := m.Namespace + ":" + m.Name
	if err := h.runCommand(shareCtx, "share-public", h.cfg.startupBound, "share", "public", "--force-agent", "--open", "-n", selection, h.target.target()); err != nil {
		return "", err
	}
	detail, err := awaitShare(shareCtx, h.cfg, h.agentRPC, h.agent, h.recordToken)
	if err != nil {
		return "", err
	}
	h.tokens = append(h.tokens, detail.GetToken())
	view, err := h.readView(ctx)
	if err != nil {
		return "", err
	}
	if err := requireName(view, m, detail.GetToken()); err != nil {
		return "", err
	}
	return fmt.Sprintf("share '%v' at '%v'", detail.GetToken(), strings.Join(detail.GetFrontendEndpoint(), ", ")), nil
}

func (h *harness) awaitReplay(ctx context.Context) (string, error) {
	m := h.st.snapshot()
	replayCtx, cancel := context.WithTimeout(ctx, h.cfg.startupBound)
	defer cancel()
	detail, err := awaitShare(replayCtx, h.cfg, h.agentRPC, h.agent, h.recordToken)
	if err != nil {
		return "", err
	}
	token := detail.GetToken()
	if slices.Contains(h.tokens, token) {
		return "", fmt.Errorf("replayed share reuses token '%v'; expected a new token", token)
	}
	h.tokens = append(h.tokens, token)
	view, err := h.readView(ctx)
	if err != nil {
		return "", err
	}
	if err := requireAbsent(view, h.tokens[:len(h.tokens)-1]); err != nil {
		return "", err
	}
	if err := requireName(view, m, token); err != nil {
		return "", err
	}
	return fmt.Sprintf("replacement share '%v' replaced '%v'", token, strings.Join(h.tokens[:len(h.tokens)-1], ", ")), nil
}

func (h *harness) checkPublic(ctx context.Context) (string, error) {
	publicCtx, cancel := context.WithTimeout(ctx, h.cfg.publicBound)
	defer cancel()
	callCtx, cancelCall := context.WithTimeout(publicCtx, h.cfg.requestBound)
	resp, err := h.agentRPC.Status(callCtx, &agentGrpc.StatusRequest{})
	cancelCall()
	if err != nil {
		return "", fmt.Errorf("agent status failed: %w", err)
	}
	detail, settled, err := activeShare(resp)
	if err != nil {
		return "", err
	}
	if !settled || detail.GetToken() != h.tokens[len(h.tokens)-1] {
		return "", fmt.Errorf("share '%v' is no longer the single active share (%v)", h.tokens[len(h.tokens)-1], describeStatus(resp))
	}
	var urls []string
	for _, endpoint := range detail.GetFrontendEndpoint() {
		url, err := awaitBody(publicCtx, h.cfg, endpoint, h.target)
		if err != nil {
			return "", err
		}
		urls = append(urls, url)
	}
	return fmt.Sprintf("exact body from '%v'", strings.Join(urls, ", ")), nil
}

// stopAgent interrupts the agent's process group as a terminal would. a timeout is a failure recorded before
// cleanup force-terminates the processes.
func (h *harness) stopAgent(ctx context.Context) (string, error) {
	agent := h.agent
	if h.agentConn != nil {
		_ = h.agentConn.Close()
		h.agentConn, h.agentRPC = nil, nil
	}
	started := time.Now()
	if err := agent.interrupt(); err != nil {
		return "", fmt.Errorf("unable to interrupt agent pid %d: %w", agent.pid(), err)
	}
	stopCtx, cancel := context.WithTimeout(ctx, h.cfg.stopBound)
	defer cancel()
	select {
	case <-agent.done:
	case <-stopCtx.Done():
		return "", fmt.Errorf("agent pid %d did not exit within %v of SIGINT: %w", agent.pid(), h.cfg.stopBound, stopCtx.Err())
	}
	if err := h.markReaped(agent.pid()); err != nil {
		return "", err
	}
	if code := agent.exitCode(); code != 0 {
		return "", fmt.Errorf("agent pid %d exited with status %d after SIGINT; see '%v'", agent.pid(), code, h.rep.logPath(fmt.Sprintf("agent-%d", h.starts)))
	}
	return fmt.Sprintf("exit 0 after %v", time.Since(started).Round(time.Millisecond)), nil
}

func (h *harness) checkStopped(ctx context.Context) (string, error) {
	m := h.st.snapshot()
	view, err := h.readView(ctx)
	if err != nil {
		return "", err
	}
	if view.environment(m.EnvZID) == nil {
		return "", fmt.Errorf("environment '%v' is missing from the account overview; absence cannot be proven", m.EnvZID)
	}
	if err := requireAbsent(view, h.tokens); err != nil {
		return "", err
	}
	if err := requireName(view, m, ""); err != nil {
		return "", err
	}
	return fmt.Sprintf("'%v' absent; '%v:%v' reserved and unbound", strings.Join(h.tokens, ", "), m.Namespace, m.Name), nil
}

func requireAbsent(view *accountView, tokens []string) error {
	for _, token := range tokens {
		if locations := view.tokenLocations(token); len(locations) > 0 {
			return fmt.Errorf("share '%v' is still present: %v", token, strings.Join(locations, "; "))
		}
	}
	return nil
}

// requireName requires exactly one listing of the run's name, reserved and bound to the given token (or unbound).
func requireName(view *accountView, m manifest, token string) error {
	matches := view.matchingNames(m.Namespace, m.Name)
	if len(matches) != 1 {
		return fmt.Errorf("name '%v:%v' is listed %d times; expected once", m.Namespace, m.Name, len(matches))
	}
	if !matches[0].Reserved {
		return fmt.Errorf("name '%v:%v' is not reserved", m.Namespace, m.Name)
	}
	if matches[0].ShareToken != token {
		if token == "" {
			return fmt.Errorf("name '%v:%v' is still bound to '%v'", m.Namespace, m.Name, matches[0].ShareToken)
		}
		return fmt.Errorf("name '%v:%v' is bound to '%v'; expected '%v'", m.Namespace, m.Name, matches[0].ShareToken, token)
	}
	return nil
}

// readView retries only failed reads, a bounded number of times; a successful read is never retried to obtain a
// different answer.
func (h *harness) readView(ctx context.Context) (*accountView, error) {
	var last error
	for attempt := 0; attempt < 3; attempt++ {
		view, err := h.ctrl.view(ctx)
		if err == nil {
			return view, nil
		}
		last = err
		timer := time.NewTimer(h.cfg.pollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, errors.Join(last, ctx.Err())
		case <-timer.C:
		}
	}
	return nil, last
}

func (h *harness) recordToken(token string) error {
	return h.st.addShareToken(token)
}

func (h *harness) markReaped(pid int) error {
	return h.st.update(func(m *manifest) {
		for i := range m.Processes {
			if m.Processes[i].PID == pid {
				m.Processes[i].Reaped = true
			}
		}
	})
}

// childEnv passes the runner's environment to the binary under test, without the account token and with the
// run's api endpoint.
func (h *harness) childEnv() []string {
	env := slices.DeleteFunc(os.Environ(), func(kv string) bool {
		return strings.HasPrefix(kv, accountTokenEnv+"=") || strings.HasPrefix(kv, "ZROK2_API_ENDPOINT=")
	})
	return append(env, "ZROK2_API_ENDPOINT="+h.st.snapshot().APIEndpoint)
}

// runCommand runs one short-lived binary invocation within its bound. its arguments, which can include the
// account token, never appear in errors or logs.
func (h *harness) runCommand(ctx context.Context, label string, bound time.Duration, args ...string) error {
	c, err := startChild(h.cfg.binary, args, h.childEnv(), h.rep.log(label))
	if err != nil {
		return fmt.Errorf("unable to start '%v': %w", label, err)
	}
	cmdCtx, cancel := context.WithTimeout(ctx, bound)
	defer cancel()
	select {
	case <-c.done:
	case <-cmdCtx.Done():
		return errors.Join(fmt.Errorf("'%v' did not complete: %w", label, cmdCtx.Err()), c.kill(h.cfg.killBound))
	}
	if code := c.exitCode(); code != 0 {
		return fmt.Errorf("'%v' exited with status %d; see '%v'", label, code, h.rep.logPath(label))
	}
	return nil
}
