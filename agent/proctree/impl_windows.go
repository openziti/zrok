//go:build windows

package proctree

import (
	"os/exec"
	"sync"

	"github.com/kolesnikovae/go-winjob"
	"golang.org/x/sys/windows"
)

var job *winjob.JobObject
var jobMu sync.Mutex

func Init(name string) error {
	jobMu.Lock()
	defer jobMu.Unlock()
	if job == nil {
		var err error
		job, err = winjob.Create(name, winjob.LimitKillOnJobClose, winjob.LimitBreakawayOK)
		return err
	}
	return nil
}

func configureChild(cmd *exec.Cmd, _ bool) error {
	if err := Init("zrok Agent"); err != nil {
		return err
	}
	cmd.SysProcAttr = &windows.SysProcAttr{CreationFlags: windows.CREATE_SUSPENDED}
	return nil
}

func finishStart(cmd *exec.Cmd) error {
	if err := job.Assign(cmd.Process); err != nil {
		return err
	}
	return winjob.ResumeProcess(cmd.Process.Pid)
}

func killChild(c *Child) error { return c.cmd.Process.Kill() }
