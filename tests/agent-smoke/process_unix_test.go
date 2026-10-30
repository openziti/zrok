//go:build agent_smoke && unix

package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"

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
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "ps"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
}

// alive reports whether a pid still names a process.
func alive(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}

// assertPrivateMode requires owner-only permissions on a retained file.
func assertPrivateMode(t *testing.T, info fs.FileInfo, path string) {
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm(), path)
}
