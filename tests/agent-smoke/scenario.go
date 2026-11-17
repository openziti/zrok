//go:build agent_smoke

package main

import (
	"context"
	"errors"
	"fmt"
	"net"
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

	// private is the token of the private share the access steps use. held is a local port this process occupies
	// so an access cannot bind it. collided is the failure id of the access that tried, while the agent holds it.
	private  string
	held     net.Listener
	collided string
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
	if !executableFile(cfg.binary, info) {
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
		{"private share through agent", h.sharePrivate},
		{"access on an occupied port", h.collideAccess},
		{"release the access", h.releaseAccess},
		{"release the private share", h.releasePrivate},
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
	args := []string{"agent", "start"}
	agent, err := startChild(h.cfg.binary, args, h.childEnv(), h.rep.log(label))
	if err != nil {
		return "", fmt.Errorf("unable to start agent: %w", err)
	}
	h.agent = agent
	record := identify(ctx, agent.pid(), h.cfg.binary, args)
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

// sharePrivate creates the private share the access steps need, through the agent.
func (h *harness) sharePrivate(ctx context.Context) (string, error) {
	shareCtx, cancel := context.WithTimeout(ctx, h.cfg.startupBound)
	defer cancel()
	if err := h.runCommand(shareCtx, "share-private", h.cfg.startupBound, "share", "private", "--force-agent", h.target.target()); err != nil {
		return "", err
	}
	detail, err := awaitPrivateShare(shareCtx, h.cfg, h.agentRPC, h.agent, h.recordToken)
	if err != nil {
		return "", err
	}
	h.private = detail.GetToken()
	view, err := h.readView(ctx)
	if err != nil {
		return "", err
	}
	if len(view.tokenLocations(h.private)) == 0 {
		return "", fmt.Errorf("private share '%v' is missing from the account overview", h.private)
	}
	return fmt.Sprintf("private share '%v'", h.private), nil
}

// collideAccess asks the agent for an access bound to a port this process already occupies. the access worker
// reports ready before it binds; the bind then fails, the worker deletes its own access and exits, and the
// agent's cleanup asks the controller to delete an access that is already gone. an agent that cannot finish that
// cleanup leaves the access 'failed' and unreleasable until it restarts, which is the defect this step catches.
// the command's own exit status depends on whether the worker reported ready first, so the agent's state decides.
func (h *harness) collideAccess(ctx context.Context) (string, error) {
	held, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("unable to occupy a local port: %w", err)
	}
	h.held = held
	addr := held.Addr().String()
	accessCtx, cancel := context.WithTimeout(ctx, h.cfg.startupBound)
	defer cancel()
	code, err := h.runCommandCode(accessCtx, "access-private", h.cfg.startupBound, "access", "private", "--force-agent", "-b", addr, h.private)
	if err != nil {
		return "", err
	}
	access, err := awaitCollision(accessCtx, h.cfg, h.agentRPC, h.agent)
	if err != nil {
		return "", err
	}
	if access == nil {
		return fmt.Sprintf("command exited %d; the agent holds no access for occupied '%v'", code, addr), nil
	}
	h.collided = access.GetFailure().GetId()
	return fmt.Sprintf("command exited %d; access is '%v' as '%v' after failing to bind occupied '%v'", code, access.GetStatus(), h.collided, addr), nil
}

// releaseAccess releases whatever the agent still holds for the collided access, then requires that neither the
// agent nor the account overview lists an access to the private share. the port stays occupied until the agent
// holds no access, so a retry cannot bind it in between.
func (h *harness) releaseAccess(ctx context.Context) (string, error) {
	released := "nothing left to release"
	if h.collided != "" {
		if err := h.runCommand(ctx, "release-access", h.cfg.commandBound, "agent", "release", "access", h.collided); err != nil {
			return "", err
		}
		released = fmt.Sprintf("released '%v'", h.collided)
	}
	waitCtx, cancel := context.WithTimeout(ctx, h.cfg.commandBound)
	defer cancel()
	if err := awaitNoAccess(waitCtx, h.cfg, h.agentRPC, h.agent); err != nil {
		return "", err
	}
	h.releaseHeld()
	view, err := h.readView(ctx)
	if err != nil {
		return "", err
	}
	if locations := view.accessLocations(h.private); len(locations) > 0 {
		return "", fmt.Errorf("an access to '%v' is still present: %v", h.private, strings.Join(locations, "; "))
	}
	return released + "; no access listed by the agent or the account overview", nil
}

// releasePrivate releases the private share and requires the agent to hold nothing and the share to be gone, so
// the public share steps start from the state they always did.
func (h *harness) releasePrivate(ctx context.Context) (string, error) {
	if err := h.runCommand(ctx, "release-share", h.cfg.commandBound, "agent", "release", "share", h.private); err != nil {
		return "", err
	}
	waitCtx, cancel := context.WithTimeout(ctx, h.cfg.commandBound)
	defer cancel()
	if err := awaitIdle(waitCtx, h.cfg, h.agentRPC, h.agent); err != nil {
		return "", err
	}
	view, err := h.readView(ctx)
	if err != nil {
		return "", err
	}
	if err := requireAbsent(view, []string{h.private}); err != nil {
		return "", err
	}
	return fmt.Sprintf("'%v' absent; the agent holds no shares or accesses", h.private), nil
}

// releaseHeld frees the occupied port. it is safe to repeat.
func (h *harness) releaseHeld() {
	if h.held != nil {
		_ = h.held.Close()
		h.held = nil
	}
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

// stopAgent delivers the platform's graceful-stop event to the agent's process group: SIGINT as a terminal's
// Ctrl-C would on POSIX, CTRL_BREAK_EVENT on Windows. a timeout is a failure recorded before cleanup
// force-terminates the processes.
func (h *harness) stopAgent(ctx context.Context) (string, error) {
	agent := h.agent
	if h.agentConn != nil {
		_ = h.agentConn.Close()
		h.agentConn, h.agentRPC = nil, nil
	}
	started := time.Now()
	h.rep.stopEvent(stopDescription())
	if err := agent.interrupt(); err != nil {
		return "", fmt.Errorf("unable to deliver %v to agent pid %d: %w", stopEvent, agent.pid(), err)
	}
	stopCtx, cancel := context.WithTimeout(ctx, h.cfg.stopBound)
	defer cancel()
	select {
	case <-agent.done:
	case <-stopCtx.Done():
		return "", fmt.Errorf("agent pid %d did not exit within %v of %v: %w", agent.pid(), h.cfg.stopBound, stopEvent, stopCtx.Err())
	}
	if err := h.markReaped(agent.pid()); err != nil {
		return "", err
	}
	if code := agent.exitCode(); code != 0 {
		return "", fmt.Errorf("agent pid %d exited with status %d after %v; see '%v'", agent.pid(), code, stopEvent, h.rep.logPath(fmt.Sprintf("agent-%d", h.starts)))
	}
	return fmt.Sprintf("exit 0 %v after %v", time.Since(started).Round(time.Millisecond), stopEvent), nil
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
	code, err := h.runCommandCode(ctx, label, bound, args...)
	if err != nil {
		return err
	}
	if code != 0 {
		return fmt.Errorf("'%v' exited with status %d; see '%v'", label, code, h.rep.logPath(label))
	}
	return nil
}

// runCommandCode is runCommand for an invocation whose exit status is an observation, not a requirement. it
// fails only when the binary cannot start or does not complete within its bound.
func (h *harness) runCommandCode(ctx context.Context, label string, bound time.Duration, args ...string) (int, error) {
	c, err := startChild(h.cfg.binary, args, h.childEnv(), h.rep.log(label))
	if err != nil {
		return 0, fmt.Errorf("unable to start '%v': %w", label, err)
	}
	cmdCtx, cancel := context.WithTimeout(ctx, bound)
	defer cancel()
	select {
	case <-c.done:
	case <-cmdCtx.Done():
		_, err := c.kill(h.cfg.killBound)
		return 0, errors.Join(fmt.Errorf("'%v' did not complete: %w", label, cmdCtx.Err()), err)
	}
	return c.exitCode(), nil
}
