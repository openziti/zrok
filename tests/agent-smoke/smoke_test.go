//go:build agent_smoke && unix

package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/michaelquigley/df/dd"
	"github.com/openziti/zrok/v2/build"
	"github.com/openziti/zrok/v2/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAccountToken = "acct-secret-0123456789"

func TestMain(m *testing.M) {
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
	// a short home keeps the agent socket path inside the unix socket limit
	home, err := os.MkdirTemp("", "smk")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(home) })
	t.Setenv("HOME", home)
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
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm(), path)
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
		{name: "nonzero shutdown", modes: []string{"exit-nonzero"}, step: "interrupt agent", text: "exited with status 1 after SIGINT"},
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
	start := func() *exec.Cmd {
		cmd := exec.Command("sleep", "30")
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		require.NoError(t, cmd.Start())
		t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait() })
		return cmd
	}
	matching, reused := start(), start()
	started, command, ok, err := processIdentity(context.Background(), matching.Process.Pid)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "sleep 30", command)

	_, err = newRunState(w.cfg.stateDir, manifest{Processes: []processRecord{
		{PID: matching.Process.Pid, Started: started, Command: command},
		{PID: reused.Process.Pid, Started: "Thu Jan  1 00:00:00 1970", Command: command},
	}})
	require.NoError(t, err)
	require.Equal(t, 0, w.cleanup())

	done := make(chan error, 1)
	go func() { done <- matching.Wait() }()
	select {
	case err := <-done:
		assert.Error(t, err, "the matching process should have been killed")
	case <-time.After(5 * time.Second):
		t.Fatal("the matching recorded process was not killed")
	}
	assert.True(t, alive(reused.Process.Pid), "a pid with a different start time must not be signalled")
	detail := w.summary("cleanup").Cleanup.Detail
	assert.Contains(t, detail, "is now a different process; not signalled")
	for _, record := range w.manifest().Processes {
		assert.True(t, record.Reaped)
	}
}

// a recorded process whose termination cleanup cannot confirm fails cleanup even when every controller resource is
// removed, and its record stays for the next cleanup, which resolves it once ps answers truthfully.
func TestUnresolvedProcessFailsCleanup(t *testing.T) {
	for _, tc := range []struct {
		name     string
		identity bool
		expect   string
	}{
		{"inspection fails", false, "unable to inspect recorded agent pid"},
		{"termination unconfirmed", true, "not confirmed terminated"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newWorld(t)
			w.cfg.killBound = time.Second
			cmd := exec.Command("sleep", "30")
			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			require.NoError(t, cmd.Start())
			t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait() })
			pid := cmd.Process.Pid
			started, command, ok, err := processIdentity(context.Background(), pid)
			require.NoError(t, err)
			require.True(t, ok)

			m := newManifest(w.cfg)
			m.EnvAttempted, m.EnvZID = true, "envproc"
			m.Processes = []processRecord{{PID: pid, Started: started, Command: command}}
			w.ctrl.addEnv(m.EnvZID, m.Description)
			_, err = newRunState(w.cfg.stateDir, m)
			require.NoError(t, err)

			path := os.Getenv("PATH")
			if tc.identity {
				fakePS(t, started, command)
			} else {
				fakePS(t, "", "")
			}
			require.Equal(t, 1, w.cleanup())
			s := w.summary("cleanup")
			assert.Equal(t, "fail", s.Result)
			assert.Equal(t, "fail", s.Cleanup.Result)
			assert.Contains(t, s.Cleanup.Detail, tc.expect)
			assert.Contains(t, s.Cleanup.Detail, strconv.Itoa(pid))
			assert.True(t, w.ctrl.empty(), "controller cleanup continues past a process failure")
			assert.False(t, w.manifest().Processes[0].Reaped)
			if !tc.identity {
				assert.True(t, alive(pid), "a process whose identity was not inspected must not be signalled")
			}

			t.Setenv("PATH", path)
			require.Equal(t, 0, w.cleanup())
			assert.Equal(t, "pass", w.summary("cleanup").Result)
			assert.True(t, w.manifest().Processes[0].Reaped)
			done := make(chan error, 1)
			go func() { done <- cmd.Wait() }()
			select {
			case err := <-done:
				assert.Error(t, err, "the recorded process should have been killed")
			case <-time.After(5 * time.Second):
				t.Fatal("the recorded process was not killed")
			}
		})
	}
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
