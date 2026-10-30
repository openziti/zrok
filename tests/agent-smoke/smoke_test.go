//go:build agent_smoke

package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/michaelquigley/df/dd"
	"github.com/openziti/zrok/v2/agent/proctree"
	"github.com/openziti/zrok/v2/build"
	"github.com/openziti/zrok/v2/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAccountToken = "acct-secret-0123456789"

func TestMain(m *testing.M) {
	// helper modes come first: a helper started from a test inherits the fake binary's environment
	if mode := os.Getenv(helperEnv); mode != "" {
		os.Exit(helperProcess(mode))
	}
	if os.Getenv(fakeCLIEnv) == "1" {
		os.Exit(fakeCLI(os.Args[1:]))
	}
	os.Exit(m.Run())
}

type world struct {
	t    *testing.T
	ctrl *fakeController
	cfg  config
	home string
}

func newWorld(t *testing.T, modes ...string) *world {
	// a short home keeps the agent socket path inside the unix socket limit. os.UserHomeDir reads HOME on unix and
	// USERPROFILE on windows; setting both points the fake binary at it on either.
	home, err := os.MkdirTemp("", "smk")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(home) })
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv(fakeCLIEnv, "1")
	t.Setenv(fakeModeEnv, strings.Join(modes, ","))
	version, hash := build.Version, build.Hash
	build.Version, build.Hash = "v2.0.0-smoke", "0123abcd"
	t.Cleanup(func() { build.Version, build.Hash = version, hash })
	t.Setenv(fakeBuildEnv, build.String())

	binary, err := filepath.Abs(os.Args[0])
	require.NoError(t, err)
	ctrl := newFakeController(t, testAccountToken)
	cfg := defaultConfig()
	cfg.binary = binary
	cfg.stateDir = t.TempDir()
	cfg.apiEndpoint = ctrl.api.URL
	cfg.accountToken = testAccountToken
	cfg.runID, cfg.runAttempt = "4242", "1"
	cfg.scenarioBound = time.Minute
	cfg.startupBound = 20 * time.Second
	cfg.stopBound = 10 * time.Second
	cfg.requestBound = 5 * time.Second
	cfg.commandBound = 20 * time.Second
	cfg.publicBound = 5 * time.Second
	cfg.cleanupBound = 30 * time.Second
	cfg.killBound = 5 * time.Second
	cfg.pollInterval = 50 * time.Millisecond
	return &world{t: t, ctrl: ctrl, cfg: cfg, home: home}
}

func (w *world) run() int {
	return runMode(context.Background(), context.Background(), w.cfg)
}

func (w *world) cleanup() int {
	return cleanupMode(context.Background(), w.cfg)
}

func (w *world) summary(mode string) summary {
	var s summary
	require.NoError(w.t, dd.BindFromJSON(&s, filepath.Join(w.cfg.stateDir, reportDir, mode+"-summary.json")))
	return s
}

func (w *world) manifest() manifest {
	st, err := loadRunState(w.cfg.stateDir)
	require.NoError(w.t, err)
	return st.snapshot()
}

func (w *world) requireFailure(step, text string) summary {
	s := w.summary("run")
	assert.Equal(w.t, "fail", s.Result)
	assert.True(w.t, strings.HasPrefix(s.Failure, step+": "), "failure '%v' should come from step '%v'", s.Failure, step)
	assert.Contains(w.t, s.Failure, text)
	return s
}

// requireCleaned checks that cleanup passed, the controller holds nothing, and no recorded agent survives.
func (w *world) requireCleaned(s summary) {
	require.NotNil(w.t, s.Cleanup)
	assert.Equal(w.t, "pass", s.Cleanup.Result, s.Cleanup.Detail)
	assert.True(w.t, w.ctrl.empty(), "controller should hold no resources")
	for _, record := range w.manifest().Processes {
		assert.True(w.t, record.Reaped)
		assert.False(w.t, alive(record.PID), "agent pid %d survived", record.PID)
	}
}

func TestScenarioPasses(t *testing.T) {
	w := newWorld(t)
	require.Equal(t, 0, w.run())

	s := w.summary("run")
	assert.Equal(t, "pass", s.Result, s.Failure)
	assert.Len(t, s.Steps, 14)
	w.requireCleaned(s)
	m := w.manifest()
	require.Len(t, m.ShareTokens, 2)
	assert.NotEqual(t, m.ShareTokens[0], m.ShareTokens[1])
	assert.Equal(t, s.Resources.ShareTokens, m.ShareTokens)
	assert.NotEmpty(t, m.EnvZID)
	assert.True(t, m.NameReserved)
	assert.Equal(t, 1, w.ctrl.callCount("POST /disable"))
	assert.Equal(t, 1, w.ctrl.callCount("DELETE /share/name"))

	// the token reaches the fake binary's output, but no retained file carries it
	var logs int
	require.NoError(t, filepath.WalkDir(w.cfg.stateDir, func(path string, entry os.DirEntry, err error) error {
		require.NoError(t, err)
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		require.NoError(t, err)
		assertPrivateMode(t, info, path)
		data, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.NotContains(t, string(data), testAccountToken, path)
		if strings.Contains(string(data), "enabling with account token '[redacted]'") {
			logs++
		}
		return nil
	}))
	assert.Equal(t, 1, logs)
}

func TestScenarioFailures(t *testing.T) {
	cases := []struct {
		name  string
		modes []string
		setup func(*world)
		step  string
		text  string
	}{
		{name: "old token persists", modes: []string{"leak-token"}, step: "controller state after stop", text: "is still present"},
		{name: "name lost", modes: []string{"drop-name"}, step: "controller state after stop", text: "listed 0 times"},
		{name: "name left bound", modes: []string{"keep-binding"}, step: "controller state after stop", text: "listed name 'public:"},
		{name: "replacement reuses token", modes: []string{"same-token"}, step: "replayed share", text: "reuses token"},
		{name: "wrong public body", setup: func(w *world) { w.ctrl.wrongBody = true; w.cfg.publicBound = 500 * time.Millisecond }, step: "public response", text: "did not return the expected body"},
		{name: "nonzero shutdown", modes: []string{"exit-nonzero"}, step: "interrupt agent", text: "exited with status 1 after " + stopEvent},
		{name: "hung shutdown", modes: []string{"hang"}, setup: func(w *world) { w.cfg.stopBound = 500 * time.Millisecond }, step: "interrupt agent", text: "did not exit within"},
		{name: "failed controller read", setup: func(w *world) { w.ctrl.failOnStop = 3 }, step: "controller state after stop", text: "account overview failed"},
		{name: "build mismatch", setup: func(w *world) { w.t.Setenv(fakeBuildEnv, "v2.0.0 [other]") }, step: "start agent", text: "expected 'v2.0.0-smoke [0123abcd]'"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := newWorld(t, tc.modes...)
			if tc.setup != nil {
				tc.setup(w)
			}
			require.Equal(t, 1, w.run())
			w.requireCleaned(w.requireFailure(tc.step, tc.text))
		})
	}
}

func TestLostEnableIsReconciledByDescription(t *testing.T) {
	w := newWorld(t, "enable-lost")
	require.Equal(t, 1, w.run())
	w.requireCleaned(w.requireFailure("enable environment", "exited"))
	m := w.manifest()
	assert.NotEmpty(t, m.EnvZID)
	assert.Contains(t, w.summary("run").Cleanup.Detail, "reconciled environment '"+m.EnvZID+"' by description")
}

func TestVersionRefusalIsAFailedPrerequisite(t *testing.T) {
	w := newWorld(t)
	w.ctrl.refuseVersion = true
	require.Equal(t, 1, w.run())
	s := w.requireFailure("controller version check", "does not match any accepted patterns")
	assert.Equal(t, "pass", s.Cleanup.Result)
	assert.Equal(t, 0, w.ctrl.callCount("POST /fake/enable"))
}

func TestNameConflictIsNotAdopted(t *testing.T) {
	w := newWorld(t)
	original := identitySuffix
	identitySuffix = func() string { return "cafe0001" }
	t.Cleanup(func() { identitySuffix = original })
	w.ctrl.addName("public", "agentsmoke-4242-1-cafe0001")

	require.Equal(t, 1, w.run())
	s := w.requireFailure("reserve name", "not adopted or deleted")
	assert.Equal(t, "pass", s.Cleanup.Result, s.Cleanup.Detail)
	assert.Equal(t, 0, w.ctrl.callCount("DELETE /share/name"))
	assert.Len(t, w.ctrl.names, 1)
	assert.Empty(t, w.ctrl.envs)
	assert.True(t, w.manifest().NameForeign)
}

func TestUnverifiableCleanupFailsWithIdentifiers(t *testing.T) {
	w := newWorld(t)
	w.ctrl.failOnStop = 1000
	require.Equal(t, 1, w.run())
	s := w.requireFailure("controller state after stop", "account overview failed")
	require.NotNil(t, s.Cleanup)
	assert.Equal(t, "fail", s.Cleanup.Result)
	m := w.manifest()
	assert.Contains(t, s.Cleanup.Detail, "environment '"+m.EnvZID+"'")
	assert.Contains(t, s.Cleanup.Detail, "share '"+m.ShareTokens[0]+"'")
	assert.Contains(t, s.Cleanup.Detail, "name 'public:"+m.Name+"'")

	// the workflow's fallback invocation completes the removal once reads recover
	w.ctrl.mu.Lock()
	w.ctrl.failReads, w.ctrl.failOnStop = 0, 0
	w.ctrl.mu.Unlock()
	require.Equal(t, 0, w.cleanup())
	w.requireCleaned(w.summary("cleanup"))
}

func TestCleanupIsRepeatable(t *testing.T) {
	w := newWorld(t, "leak-token")
	require.Equal(t, 1, w.run())
	w.requireCleaned(w.requireFailure("controller state after stop", "is still present"))
	disables := w.ctrl.callCount("POST /disable")

	for i := 0; i < 2; i++ {
		require.Equal(t, 0, w.cleanup())
		s := w.summary("cleanup")
		assert.Equal(t, "pass", s.Result)
		assert.Contains(t, s.Cleanup.Detail, "verified absence")
	}
	assert.Equal(t, disables, w.ctrl.callCount("POST /disable"))
	assert.True(t, w.ctrl.empty())
}

func TestCleanupWithoutManifest(t *testing.T) {
	w := newWorld(t)
	w.cfg.accountToken = ""
	require.Equal(t, 0, w.cleanup())
	assert.Contains(t, w.summary("cleanup").Cleanup.Detail, "no manifest")
}

func TestCleanupRequiresCredentialForRecordedResources(t *testing.T) {
	w := newWorld(t, "leak-token")
	require.Equal(t, 1, w.run())
	w.cfg.accountToken = ""
	require.Equal(t, 1, w.cleanup())
	assert.Contains(t, w.summary("cleanup").Cleanup.Detail, accountTokenEnv)
}

func TestPreflightRefusals(t *testing.T) {
	t.Run("missing credential", func(t *testing.T) {
		w := newWorld(t)
		w.cfg.accountToken = ""
		require.Equal(t, 1, w.run())
		assert.Contains(t, w.summary("run").Failure, accountTokenEnv)
		assert.Empty(t, w.ctrl.calls)
	})
	t.Run("existing environment", func(t *testing.T) {
		w := newWorld(t)
		require.NoError(t, os.Mkdir(filepath.Join(w.home, ".zrok2"), 0o700))
		require.Equal(t, 1, w.run())
		assert.Contains(t, w.summary("run").Failure, "never adopts an existing environment")
		assert.Empty(t, w.ctrl.calls)
		_, err := os.Stat(filepath.Join(w.cfg.stateDir, manifestName))
		assert.ErrorIs(t, err, os.ErrNotExist)
	})
	t.Run("developer build", func(t *testing.T) {
		w := newWorld(t)
		build.Version = ""
		require.Equal(t, 1, w.run())
		assert.Contains(t, w.summary("run").Failure, "build.Version and build.Hash")
		assert.Empty(t, w.ctrl.calls)
	})
	t.Run("reused state directory", func(t *testing.T) {
		w := newWorld(t)
		require.NoError(t, os.WriteFile(filepath.Join(w.cfg.stateDir, manifestName), []byte("{}"), 0o600))
		require.Equal(t, 1, w.run())
		assert.Contains(t, w.summary("run").Failure, "already exists")
		assert.Empty(t, w.ctrl.calls)
	})
}

func TestRecordedProcessGuard(t *testing.T) {
	w := newWorld(t)
	matching, reused := startRecordable(t), startRecordable(t)
	matchingRecord := recordOf(t, matching)
	reusedRecord := recordOf(t, reused)
	reusedRecord.Started = "0"

	_, err := newRunState(w.cfg.stateDir, manifest{Processes: []processRecord{matchingRecord, reusedRecord}})
	require.NoError(t, err)
	require.Equal(t, 0, w.cleanup())

	requireKilled(t, matching, "the matching recorded process")
	assert.True(t, alive(reused.Process.Pid), "a pid with a different start time must not be signalled")
	detail := w.summary("cleanup").Cleanup.Detail
	assert.Contains(t, detail, "is now a different process; not signalled")
	assert.Contains(t, detail, fmt.Sprintf("force-terminated recorded agent pid %d", matching.Process.Pid))
	for _, record := range w.manifest().Processes {
		assert.True(t, record.Reaped)
	}
}

// a recorded process whose termination cleanup cannot confirm fails cleanup even when every controller resource is
// removed, and its record stays for the next cleanup, which resolves it once inspection answers truthfully.
func TestUnresolvedProcessFailsCleanup(t *testing.T) {
	for _, tc := range []struct {
		name        string
		unconfirmed bool
		expect      string
	}{
		{"inspection fails", false, "unable to inspect recorded agent pid"},
		{"termination unconfirmed", true, "not confirmed terminated"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newWorld(t)
			w.cfg.killBound = time.Second
			cmd := startRecordable(t)
			pid := cmd.Process.Pid
			record := recordOf(t, cmd)

			m := newManifest(w.cfg)
			m.EnvAttempted, m.EnvZID = true, "envproc"
			m.Processes = []processRecord{record}
			w.ctrl.addEnv(m.EnvZID, m.Description)
			_, err := newRunState(w.cfg.stateDir, m)
			require.NoError(t, err)

			restore := breakInspection(t, record, tc.unconfirmed)
			require.Equal(t, 1, w.cleanup())
			s := w.summary("cleanup")
			assert.Equal(t, "fail", s.Result)
			assert.Equal(t, "fail", s.Cleanup.Result)
			assert.Contains(t, s.Cleanup.Detail, tc.expect)
			assert.Contains(t, s.Cleanup.Detail, strconv.Itoa(pid))
			assert.True(t, w.ctrl.empty(), "controller cleanup continues past a process failure")
			assert.False(t, w.manifest().Processes[0].Reaped)
			if !tc.unconfirmed {
				assert.True(t, alive(pid), "a process whose identity was not inspected must not be signalled")
			}

			restore()
			require.Equal(t, 0, w.cleanup())
			assert.Equal(t, "pass", w.summary("cleanup").Result)
			assert.True(t, w.manifest().Processes[0].Reaped)
			requireKilled(t, cmd, "the recorded process")
		})
	}
}

// a record written on another platform holds an identity this build cannot check, so its pid is never acted on
// and the record is kept.
func TestForeignPlatformRecordIsKept(t *testing.T) {
	w := newWorld(t)
	cmd := startRecordable(t)
	record := recordOf(t, cmd)
	record.Platform = posixPlatform
	if localPlatform == posixPlatform {
		record.Platform = windowsPlatform
	}
	_, err := newRunState(w.cfg.stateDir, manifest{Processes: []processRecord{record}})
	require.NoError(t, err)

	require.Equal(t, 1, w.cleanup())
	s := w.summary("cleanup")
	assert.Contains(t, s.Cleanup.Detail, fmt.Sprintf("has a '%v' identity and cannot be inspected on '%v'", record.Platform, localPlatform))
	assert.True(t, alive(cmd.Process.Pid), "a foreign record's pid must not be signalled")
	assert.False(t, w.manifest().Processes[0].Reaped)
}

func TestRecordsWithoutPlatformArePOSIX(t *testing.T) {
	assert.Equal(t, posixPlatform, recordPlatform(processRecord{PID: 1, Started: "Thu Jan  1 00:00:00 1970"}))
	assert.Equal(t, windowsPlatform, recordPlatform(processRecord{PID: 1, Platform: windowsPlatform}))
}

// the delivery the live scenario uses for a graceful stop reaches a child that installed go's interrupt handler,
// which acknowledges it and exits normally.
func TestGracefulEventDelivery(t *testing.T) {
	out := &boundedLog{}
	c := startHelperChild(t, "interrupt", out)
	awaitOutput(t, out, "ready")
	require.NoError(t, c.interrupt())
	select {
	case <-c.done:
	case <-time.After(10 * time.Second):
		t.Fatalf("helper did not exit within 10s of %v", stopEvent)
	}
	assert.Equal(t, 0, c.exitCode(), string(out.contents()))
	assert.Contains(t, string(out.contents()), "interrupted")
}

// forced termination reaches a worker started the way the agent starts its workers, in their own process group
// on POSIX and in a nested job object on Windows, not just the direct child.
func TestKillTerminatesDescendants(t *testing.T) {
	out := &boundedLog{}
	c := startHelperChild(t, "worker", out)
	grandchild := awaitGrandchild(t, out)
	require.True(t, alive(grandchild))

	killed, err := c.kill(5 * time.Second)
	require.NoError(t, err)
	assert.Contains(t, killed, fmt.Sprintf("pid %d", c.pid()))
	require.Eventually(t, func() bool { return !alive(grandchild) }, 5*time.Second, 50*time.Millisecond, "grandchild pid %d survived", grandchild)
	again, err := c.kill(5 * time.Second)
	require.NoError(t, err)
	assert.Empty(t, again, "a second kill has nothing left to terminate")
}

func TestGeneratedNamesAreValid(t *testing.T) {
	for _, run := range [][2]string{{"local", "0"}, {"18734567890", "3"}, {strings.Repeat("9", 80), "12"}, {"Weird_ID", "1"}} {
		cfg := defaultConfig()
		cfg.runID, cfg.runAttempt = run[0], run[1]
		m := newManifest(cfg)
		assert.True(t, util.IsValidNameInNamespace(m.Name, false), "name '%v'", m.Name)
		assert.True(t, strings.HasPrefix(m.Name, "agentsmoke-"))
		assert.NotEqual(t, m.Name, newManifest(cfg).Name)
	}
}

func TestEntryRejectsUsage(t *testing.T) {
	var stderr bytes.Buffer
	assert.Equal(t, 2, entry(nil, &stderr))
	assert.Equal(t, 2, entry([]string{"run", "--state-dir", t.TempDir()}, &stderr))
	assert.Equal(t, 2, entry([]string{"cleanup"}, &stderr))
	assert.Equal(t, 2, entry([]string{"sweep", "--state-dir", t.TempDir()}, &stderr))
	assert.Contains(t, stderr.String(), "usage: agent-smoke")
}

// the harness must stay out of ordinary package discovery: untagged builds, tests, installs and vet.
func TestOrdinaryDiscoveryExcludesHarness(t *testing.T) {
	gomod, err := exec.Command("go", "env", "GOMOD").Output()
	require.NoError(t, err)
	cmd := exec.Command("go", "list", "./...")
	cmd.Dir = filepath.Dir(strings.TrimSpace(string(gomod)))
	cmd.Env = append(os.Environ(), "GOFLAGS=")
	out, err := cmd.Output()
	require.NoError(t, err)
	assert.Contains(t, string(out), "github.com/openziti/zrok/v2/agent\n")
	assert.NotContains(t, string(out), "tests/agent-smoke")
}

const helperEnv = "AGENT_SMOKE_HELPER"

// helperProcess runs the test binary as a plain process for the process tests: wait idles, interrupt
// acknowledges a graceful-stop event and exits normally, worker starts a waiting grandchild through the agent's
// own worker launcher, and spawn starts one directly, exiting at once when AGENT_SMOKE_HELPER_ORPHAN is set so
// the grandchild outlives it.
func helperProcess(mode string) int {
	switch mode {
	case "wait":
		fmt.Printf("grandchild %d\n", os.Getpid())
		time.Sleep(5 * time.Minute)
		return 0
	case "interrupt":
		interrupts := make(chan os.Signal, 1)
		signal.Notify(interrupts, os.Interrupt)
		fmt.Println("ready")
		select {
		case <-interrupts:
			fmt.Println("interrupted")
			return 0
		case <-time.After(2 * time.Minute):
			return 3
		}
	case "worker":
		_ = os.Setenv(helperEnv, "wait")
		// started as the agent starts a public worker; it reports its own pid, so relay its output
		relay := func(data []byte) { _, _ = os.Stdout.Write(data) }
		if _, err := proctree.StartControlledChild(proctree.ControlledOptions{Stdout: relay, Stderr: relay}, os.Args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "unable to start worker: %v\n", err)
			return 1
		}
		time.Sleep(5 * time.Minute)
		return 0
	case "spawn":
		cmd := exec.Command(os.Args[0])
		cmd.Env = append(os.Environ(), helperEnv+"=wait")
		if err := cmd.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "unable to start grandchild: %v\n", err)
			return 1
		}
		fmt.Printf("grandchild %d\n", cmd.Process.Pid)
		if os.Getenv(helperEnv+"_ORPHAN") != "" {
			return 0
		}
		time.Sleep(5 * time.Minute)
		return 0
	}
	fmt.Fprintf(os.Stderr, "unknown helper mode '%v'\n", mode)
	return 2
}

func testBinary(t *testing.T) string {
	binary, err := filepath.Abs(os.Args[0])
	require.NoError(t, err)
	return binary
}

// startHelperChild starts a helper exactly as the harness starts the agent, and force-terminates it at test end.
func startHelperChild(t *testing.T, mode string, out *boundedLog, extraEnv ...string) *child {
	env := append(os.Environ(), helperEnv+"="+mode)
	c, err := startChild(testBinary(t), nil, append(env, extraEnv...), out)
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = c.kill(5 * time.Second) })
	return c
}

func awaitOutput(t *testing.T, out *boundedLog, text string) {
	require.Eventually(t, func() bool { return strings.Contains(string(out.contents()), text) }, 10*time.Second, 20*time.Millisecond, "helper never printed '%v'", text)
}

var grandchildLine = regexp.MustCompile(`grandchild (\d+)`)

func awaitGrandchild(t *testing.T, out *boundedLog) int {
	awaitOutput(t, out, "grandchild ")
	match := grandchildLine.FindStringSubmatch(string(out.contents()))
	require.NotNil(t, match)
	pid, err := strconv.Atoi(match[1])
	require.NoError(t, err)
	t.Cleanup(func() {
		if p, err := os.FindProcess(pid); err == nil {
			_ = p.Kill()
		}
	})
	return pid
}

// recordOf records a helper's identity as the harness records an agent's.
func recordOf(t *testing.T, cmd *exec.Cmd) processRecord {
	record := identify(context.Background(), cmd.Process.Pid, testBinary(t), nil)
	require.NotEmpty(t, record.Started, "identity of pid %d", cmd.Process.Pid)
	return record
}

func requireKilled(t *testing.T, cmd *exec.Cmd, what string) {
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		assert.Error(t, err, "%v should have been killed", what)
	case <-time.After(5 * time.Second):
		t.Fatalf("%v was not killed", what)
	}
}
