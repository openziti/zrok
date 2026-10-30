//go:build agent_smoke && windows

package main

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

// startRecordable starts a waiting helper outside any harness ownership, as an earlier invocation's agent looks
// to a later cleanup.
func startRecordable(t *testing.T) *exec.Cmd {
	cmd := exec.Command(testBinary(t))
	cmd.Env = append(os.Environ(), helperEnv+"=wait")
	require.NoError(t, cmd.Start())
	t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait() })
	return cmd
}

// breakInspection makes reading a recorded process's identity fail or, for unconfirmed, makes termination a
// no-op so the process never exits. the returned function restores the real calls.
func breakInspection(t *testing.T, _ processRecord, unconfirmed bool) func() {
	inspect, terminate := inspectProcess, terminateProcess
	if unconfirmed {
		terminateProcess = func(windows.Handle) error { return nil }
	} else {
		inspectProcess = func(windows.Handle) (string, string, error) {
			return "", "", errors.New("inspection refused by the test")
		}
	}
	restore := func() { inspectProcess, terminateProcess = inspect, terminate }
	t.Cleanup(restore)
	return restore
}

// alive reports whether a pid names a process that has not exited.
func alive(pid int) bool {
	h, err := windows.OpenProcess(windows.SYNCHRONIZE|windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	defer func() { _ = windows.CloseHandle(h) }()
	return !signalled(h)
}

// assertPrivateMode checks nothing on windows: mode bits there say nothing about the ACL. retained files live in
// the runner-owned state directory.
func assertPrivateMode(*testing.T, fs.FileInfo, string) {}

// a descendant that outlives the child, as a worker outlives a failed agent, is still in the child's job and is
// terminated with it.
func TestOrphanedDescendantIsTerminated(t *testing.T) {
	out := &boundedLog{}
	c := startHelperChild(t, "spawn", out, helperEnv+"_ORPHAN=1")
	grandchild := awaitGrandchild(t, out)
	select {
	case <-c.done:
	case <-time.After(10 * time.Second):
		t.Fatal("spawning helper did not exit")
	}
	require.Equal(t, 0, c.exitCode())
	require.True(t, alive(grandchild), "the grandchild outlives its parent")

	killed, err := c.kill(5 * time.Second)
	require.NoError(t, err)
	assert.Contains(t, killed, "processes in its job")
	require.Eventually(t, func() bool { return !alive(grandchild) }, 5*time.Second, 50*time.Millisecond, "orphaned grandchild pid %d survived", grandchild)
}

// a recorded process from an earlier invocation is terminated with its descendants, through handles held across
// the identity check.
func TestRecordedProcessTreeIsTerminated(t *testing.T) {
	w := newWorld(t)
	cmd := exec.Command(testBinary(t))
	cmd.Env = append(os.Environ(), helperEnv+"=spawn")
	out := &boundedLog{}
	cmd.Stdout = out
	require.NoError(t, cmd.Start())
	t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait() })
	grandchild := awaitGrandchild(t, out)

	_, err := newRunState(w.cfg.stateDir, manifest{Processes: []processRecord{recordOf(t, cmd)}})
	require.NoError(t, err)
	require.Equal(t, 0, w.cleanup())
	requireKilled(t, cmd, "the recorded process")
	require.Eventually(t, func() bool { return !alive(grandchild) }, 5*time.Second, 50*time.Millisecond, "descendant pid %d survived", grandchild)
}
