//go:build !windows

package subordinate

import (
	"io"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

// inherited stdin may be a blocking descriptor. a pollable duplicate lets Close interrupt its watcher.
func workerInput(input io.ReadCloser) (io.ReadCloser, error) {
	file, ok := input.(*os.File)
	if !ok || file.SetReadDeadline(time.Time{}) == nil {
		return input, nil
	}
	conn, err := file.SyscallConn()
	if err != nil {
		return nil, err
	}
	var fd int
	var duplicateErr error
	if err := conn.Control(func(handle uintptr) { fd, duplicateErr = unix.FcntlInt(handle, unix.F_DUPFD_CLOEXEC, 0) }); err != nil {
		return nil, err
	}
	if duplicateErr != nil {
		return nil, duplicateErr
	}
	if err := unix.SetNonblock(fd, true); err != nil {
		_ = unix.Close(fd)
		return nil, err
	}
	return os.NewFile(uintptr(fd), "worker control input"), nil
}
