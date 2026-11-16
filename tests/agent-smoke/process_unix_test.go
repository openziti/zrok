//go:build agent_smoke && unix

package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// startRecordable starts a waiting helper in its own process group, as the harness records an agent, outside
// any harness ownership.
func startRecordable(t *testing.T) *exec.Cmd {
	cmd := exec.Command(testBinary(t))
	cmd.Env = append(os.Environ(), helperEnv+"=wait")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	require.NoError(t, cmd.Start())
	t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait() })
	return cmd
}

// breakInspection puts a ps on PATH that fails every inspection or, for unconfirmed, reports the recorded
// identity for any pid as a running process, so the process never appears to terminate. the returned function
// restores the real ps.
func breakInspection(t *testing.T, record processRecord, unconfirmed bool) func() {
	path := os.Getenv("PATH")
	if unconfirmed {
		fakePS(t, record.Started, record.Command)
	} else {
		fakePS(t, "", "")
	}
	return func() { t.Setenv("PATH", path) }
}

func fakePS(t *testing.T, started, command string) {
	script := "#!/bin/sh\nexit 2\n"
	if started != "" {
		script = fmt.Sprintf("#!/bin/sh\ncase \"$2\" in\nlstart=) echo '%v' ;;\nargs=) echo '%v' ;;\n*) echo S ;;\nesac\n", started, command)
	}
	installPS(t, script)
}

// installPS makes script the only ps on PATH for the rest of the test.
func installPS(t *testing.T, script string) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "ps"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
}

// the bound catching an inspection mid-run is the bound expiring, not an inspection failure. a ps that always
// outlasts the bound makes that happen on every run; left to chance, it depends on where the deadline falls among
// the polls, which is how a loaded runner came to report the killed ps as the reason.
func TestAwaitGoneDoesNotReportTheInspectionItInterrupted(t *testing.T) {
	// exec, so the kill reaches the sleeper itself and no child is left holding the output pipe open.
	installPS(t, "#!/bin/sh\nexec /bin/sleep 30\n")
	record := processRecord{PID: os.Getpid(), Platform: "posix", Started: "then", Command: "it"}
	bound := 200 * time.Millisecond
	start := time.Now()
	err := awaitGone(context.Background(), record, bound, 10*time.Millisecond)
	require.Error(t, err)
	assert.Equal(t, "not confirmed terminated within 200ms of being killed", err.Error())
	assert.Less(t, time.Since(start), 5*time.Second, "the interrupted inspection must not outlive the bound by long")
}

// an inspection that fails inside the bound is still reported with its cause, including when a later inspection
// is the one the bound interrupts.
func TestAwaitGoneReportsAnInspectionFailure(t *testing.T) {
	installPS(t, "#!/bin/sh\nexit 2\n")
	record := processRecord{PID: os.Getpid(), Platform: "posix", Started: "then", Command: "it"}
	err := awaitGone(context.Background(), record, 300*time.Millisecond, 10*time.Millisecond)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "termination not confirmed within 300ms of being killed: ")
	assert.Contains(t, err.Error(), "exit status 2")
}

// alive reports whether a pid still names a process.
func alive(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}

// assertPrivateMode requires owner-only permissions on a retained file.
func assertPrivateMode(t *testing.T, info fs.FileInfo, path string) {
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm(), path)
}
