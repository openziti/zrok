//go:build !windows

package proctree

import (
	"errors"
	"os/exec"
	"syscall"
)

func Init(_ string) error { return nil }

func configureChild(cmd *exec.Cmd, controlled bool) error {
	if controlled {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}
	return nil
}

func finishStart(_ *exec.Cmd) error { return nil }

func killChild(c *Child) error {
	if !c.controlled {
		return c.cmd.Process.Kill()
	}
	err := syscall.Kill(-c.cmd.Process.Pid, syscall.SIGKILL)
	if errors.Is(err, syscall.ESRCH) {
		return nil
	}
	return err
}
